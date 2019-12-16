## Anywhere
Anywhere is an open-source intranet pass-through service. underlying implementation borrowed from ngrok(io.Copy).

**Only Support Linux For Now, and not fully tested, maybe not stable**


### QuickStart
Suppose we have two machine.
* Server -- Extranet Server
* Agent -- Intranet Server

and we want : Server provide intranet pass-through service for Agent.
```bash
### build
On Any Server>> make
On Any Server>> ll anywhere.tar.gz

### Upload anywhere.tar.gz to Server and Agent
On Server & Agent>> mkdir /usr/local/anywhere && tar -xvf anywhere.tar.gz -C /usr/local/anywhere


### Start Anywhere Server
On Server >> cd /usr/local/anywhere
On Server >> nohup ./bin/anywhered start --wl 8.8.8.8,7.7.7.7 & #8.8.8.8,7.7.7.7 should replace by your own IP. should contains agent's public IP.

### Start Anywhere Agent
On Agent >>  cd /usr/local/anywhere
On Agent >> nohup ./bin/anywhere -i anywhere-agent-1 -s 8.8.8.8 & # 8.8.8.8 shoud replace by your own Anywhere Server IP

### Add Tunnel Config
On Server >> cd /usr/local/anywhere
On Server >> ./bin/anywhered proxy add --agent-id anywhere-agent-1 --local-ip 10.0.0.2 --local-port 80 --remote-port 80 
# local-ip, local-port should replace by your own IP.

## Access Anywhere Server:80, will be same as access 10.0.0.2:80
```

### Glossary
* Anywhere Server: An Extranet server, accept connection from Internet.
* Anywhere Agent: An Intranet agent, provide data connection to Anywhere Server.
* Data Connection: A Connection to specified Intranet target.
* Admin Connection: A long connection between Anywhere Server and Anywhere Agent.
  * Sending Heartbeat to keep alive.
  * Anywhere Server use this connection to ask for a Data Connection to specified Intranet target.

### Workflow:
* 1.Server Started with Below Tunnel Config. 
```json
{
    //Just for demo, actual config is not using json format
    "remote-port":"80",
    "agent-id":"agent-1",
    "local-ip":"10.0.0.2",
    "local-port":"80",
    "white-list":"8.8.8.8"
}
 ```
* 2.Anywhere Server start listen 0.0.0.0:80 for Internet conncection.
* 3.`Anywhere-Server:80` got a connection from Internet.
  * 3.1 check `Internet conncection` remoteAddr is in whiteList or not. not in whiteList will close it.
  * 3.2 ask `agent-1` for Data Connection to `10.0.0.2:80`(local-ip:local-port) and wait.
  * 3.3 got Data Connection, join `Internet conncection` and `Data Connection` or got `Data Connection` timeout, close `Internet conncection`



### Further Info
Just Use --help
```bash
[root@70a6f341999b anywhere]# ./bin/anywhere --help
anywhere agent - "master-master d6f5d6e2761777a25f54b9a6e1e58063463fe8af"

Usage:
  anywhere --help [flags]

Flags:
  -i, --agent-id string    anywhere agent id (default "anywhere-agent-1")
      --ca string          ca file (default "credential/ca.crt")
      --cert string        cert file (default "credential/client.crt")
  -h, --help               help for anywhere
      --key string         key file (default "credential/client.key")
  -s, --server-ip string   anywhered server address (default "127.0.0.1")
  -p, --server-port int    anywhered server port (default 1111)

[root@70a6f341999b anywhere]# ./bin/anywhered --help
anywhere server - "master-master d6f5d6e2761777a25f54b9a6e1e58063463fe8af"

Usage:
  anywhered [command]

Available Commands:
  agent       agent admin interface
  help        Help about any command
  proxy       proxy admin interface
  start       start anywhered service

Flags:
  -a, --api-port int       anywhered rest api port (default 1112)
      --ca string          ca file (default "credential/ca.crt")
      --cert string        cert file (default "credential/server.crt")
  -g, --grpc-port int      anywhered grpc port (default 1113)
  -h, --help               help for anywhered
      --key string         key file (default "credential/server.key")
  -p, --port int           anywhered serve port (default 1111)
  -s, --server-id string   anywhered server id (default "anywhered-1")

Use "anywhered [command] --help" for more information about a command.
```
