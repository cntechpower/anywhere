package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cntechpower/anywhere/agent/agent"
	"github.com/cntechpower/anywhere/model"
	tlsutil "github.com/cntechpower/anywhere/tls"
)

func writePEMFile(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		t.Fatal(err)
	}
}

type e2eCerts struct {
	caFile, serverCert, serverKey, clientCert, clientKey string
}

func generateE2ECerts(t *testing.T) e2eCerts {
	t.Helper()
	dir := t.TempDir()
	now := time.Now()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "e2e-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	mk := func(serial int64, cn string, dns []string, usages []x509.ExtKeyUsage) (string, string) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		tpl := &x509.Certificate{
			SerialNumber: big.NewInt(serial),
			Subject:      pkix.Name{CommonName: cn},
			NotBefore:    now.Add(-time.Hour),
			NotAfter:     now.Add(time.Hour),
			DNSNames:     dns,
			KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage:  usages,
		}
		der, err := x509.CreateCertificate(rand.Reader, tpl, caTpl, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		certFile := filepath.Join(dir, cn+".crt")
		keyFile := filepath.Join(dir, cn+".key")
		writePEMFile(t, certFile, "CERTIFICATE", der)
		writePEMFile(t, keyFile, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
		return certFile, keyFile
	}

	serverCert, serverKey := mk(2, "server", []string{"anywhere.local"}, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	clientCert, clientKey := mk(3, "client", nil, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	caFile := filepath.Join(dir, "ca.crt")
	writePEMFile(t, caFile, "CERTIFICATE", caDER)

	return e2eCerts{caFile: caFile, serverCert: serverCert, serverKey: serverKey, clientCert: clientCert, clientKey: clientKey}
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startEchoServer(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}()
		}
	}()
	return ln.Addr().String(), func() { _ = ln.Close() }
}

// TestEndToEndTCPProxy exercises the whole path: agent connects over mTLS,
// registers, the server asks it for a tunnel, the agent dials the local
// service, and a client connection is proxied through.
func TestEndToEndTCPProxy(t *testing.T) {
	certs := generateE2ECerts(t)
	echoAddr, stopEcho := startEchoServer(t)
	defer stopEcho()

	mainPort := freePort(t)
	proxyPort := freePort(t)

	serverCfg, err := tlsutil.ParseServerTlsConfig(certs.serverCert, certs.serverKey, certs.caFile)
	if err != nil {
		t.Fatal(err)
	}

	s, err := New("e2e-server", mainPort, &model.UserConfig{Users: []*model.User{
		{UserName: "admin", UserPass: "admin"},
	}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	s.SetCredentials(serverCfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = s.Shutdown(shutdownCtx)
	}()

	a, err := agent.New("e2e-zone", "e2e-agent", "127.0.0.1", "admin", "admin", mainPort)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SetCredentials(certs.clientCert, certs.clientKey, certs.caFile, "anywhere.local", false); err != nil {
		t.Fatal(err)
	}
	a.Start(ctx)
	defer a.Stop()

	deadline := time.Now().Add(5 * time.Second)
	for len(s.ListAgentInfo()) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("agent did not register in time")
		}
		time.Sleep(20 * time.Millisecond)
	}

	if err := s.AddProxyConfig("admin", "e2e-zone", proxyPort, echoAddr, false, "", model.ListenTypeTCP); err != nil {
		t.Fatalf("add proxy config: %v", err)
	}

	client, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort), 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(5 * time.Second))

	payload := []byte("hello-anywhere")
	if _, err := client.Write(payload); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(payload))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(buf) != string(payload) {
		t.Fatalf("got %q, want %q", string(buf), string(payload))
	}

	// Close the client so the proxy handler finishes and records the counters.
	_ = client.Close()
	deadline = time.Now().Add(3 * time.Second)
	for {
		s.refreshSummary()
		if s.GetSummary().ProxyConnectTotalCount == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("proxy connect count = %d, want 1", s.GetSummary().ProxyConnectTotalCount)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
