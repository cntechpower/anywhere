package constants

const AnywhereVersion = "0.0.2"
const DefaultTimeFormat = "2006-01-02 15:04:05.000"

// anywhered status cache
const CacheRefreshLoopTimeSeconds = 5
const CacheRefreshLogInhibition = 120 //log status generate time cost detail for every 600 (120 * 5) seconds

// anywhered config file name
const ProxyConfigFileName = "proxy.json"
const SystemConfigFIle = "anywhered.json"

// anywhered agent proxy connection
const ProxyConnBufferForEachAgent = 10
const ProxyConnGetRetryMilliseconds = 200
const ProxyConnGetMaxRetryCount = 5
const ProxyConnMaxIdleTimeout = 30
