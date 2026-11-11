package constants

const AnywhereVersion = "0.0.3"
const DefaultTimeFormat = "2006-01-02 15:04:05.000"

// anywhered status cache
const CacheRefreshLoopTimeSeconds = 5
const CacheRefreshLogInhibition = 120 //log status generate time cost detail for every 600 (120 * 5) seconds

// anywhered config file name
const ProxyConfigFileName = "proxy.json"
const SystemConfigFile = "anywhered.json"

// anywhered agent proxy connection
const ProxyConnBufferForEachAgent = 10
const ProxyConnGetRetryMilliseconds = 200
const ProxyConnGetMaxRetryCount = 5
const ProxyConnMaxIdleTimeout = 30

// ProxyMaxConnectionsPerConfig bounds concurrent client connections handled by
// one proxy config. When the limit is reached new connections are rejected
// instead of exhausting file descriptors.
const ProxyMaxConnectionsPerConfig = 4096

// ProxyConnIdleTimeoutSeconds closes proxied connections that are completely
// idle for this long. 0 disables the timeout (needed for long lived tunnels
// such as SSH sessions that may stay quiet for hours).
const ProxyConnIdleTimeoutSeconds = 0

// ProxyTokenTTLSeconds is the lifetime of the single use token that binds a
// data connection to an authenticated control connection.
const ProxyTokenTTLSeconds = 30

// AgentHeartbeatTimeoutSeconds is how long an agent may stay silent before it
// is considered dead and removed from the zone.
const AgentHeartbeatTimeoutSeconds = 300

// UDPDataChannelBuffer is the number of datagrams queued towards an agent per
// proxy config. When the queue is full datagrams are dropped (UDP semantics).
const UDPDataChannelBuffer = 1024

// UDPMaxDatagramSize is the largest UDP payload we forward.
const UDPMaxDatagramSize = 65535

// UDPSessionIdleSeconds closes a UDP session that saw no traffic for this long.
const UDPSessionIdleSeconds = 120
