package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testCerts struct {
	caFile     string
	serverCert string
	serverKey  string
	clientCert string
	clientKey  string
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
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

func generateCerts(t *testing.T, dnsNames []string) testCerts {
	t.Helper()
	dir := t.TempDir()
	now := time.Now()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "anywhere-test-ca"},
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

	mkCert := func(serial int64, cn string, dns []string, usages []x509.ExtKeyUsage) (certFile, keyFile string) {
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
		certFile = filepath.Join(dir, cn+".crt")
		keyFile = filepath.Join(dir, cn+".key")
		writePEM(t, certFile, "CERTIFICATE", der)
		writePEM(t, keyFile, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
		return certFile, keyFile
	}

	serverCert, serverKey := mkCert(2, "server", dnsNames, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	clientCert, clientKey := mkCert(3, "client", nil, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	caFile := filepath.Join(dir, "ca.crt")
	writePEM(t, caFile, "CERTIFICATE", caDER)

	return testCerts{
		caFile:     caFile,
		serverCert: serverCert,
		serverKey:  serverKey,
		clientCert: clientCert,
		clientKey:  clientKey,
	}
}

func TestParseServerTlsConfig(t *testing.T) {
	c := generateCerts(t, []string{"anywhere.local"})
	cfg, err := ParseServerTlsConfig(c.serverCert, c.serverKey, c.caFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatal("server must require client certificates")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatal("min version must be TLS 1.2")
	}
	if cfg.InsecureSkipVerify {
		t.Fatal("server config must not skip verification")
	}
}

func TestClientServerHandshake(t *testing.T) {
	c := generateCerts(t, []string{"anywhere.local"})
	serverCfg, err := ParseServerTlsConfig(c.serverCert, c.serverKey, c.caFile)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg, err := ParseClientTlsConfig(c.clientCert, c.clientKey, c.caFile, "anywhere.local", false)
	if err != nil {
		t.Fatal(err)
	}
	if clientCfg.InsecureSkipVerify {
		t.Fatal("client must verify the server by default")
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	accepted := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			accepted <- err
			return
		}
		_ = conn.(*tls.Conn).Handshake()
		accepted <- nil
		_ = conn.Close()
	}()

	raw, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	client := tls.Client(raw, clientCfg)
	if err := client.Handshake(); err != nil {
		t.Fatalf("handshake failed: %v", err)
	}
	if err := <-accepted; err != nil {
		t.Fatalf("server handshake failed: %v", err)
	}
}

func TestClientRejectsWrongServerName(t *testing.T) {
	c := generateCerts(t, []string{"anywhere.local"})
	serverCfg, err := ParseServerTlsConfig(c.serverCert, c.serverKey, c.caFile)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg, err := ParseClientTlsConfig(c.clientCert, c.clientKey, c.caFile, "wrong.name", false)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		if conn, err := ln.Accept(); err == nil {
			_ = conn.(*tls.Conn).Handshake()
			_ = conn.Close()
		}
	}()

	raw, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	client := tls.Client(raw, clientCfg)
	if err := client.Handshake(); err == nil {
		t.Fatal("expected handshake to fail with a wrong server name")
	}
}
