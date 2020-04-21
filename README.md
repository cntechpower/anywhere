## Anywhere
Anywhere is an open-source intranet pass-through service. like [ngrok](https://github.com/inconshreveable/ngrok)


### Quickstart
Suppose we have two machine.
* Server -- Extranet Server had public IP, Suppose public ip is `1.2.3.4`
* Agent -- Intranet Server

```bash


### Upload anywhere.tar.gz to Server and Agent
On Server & Agent # wget http://internal.cntechpower.com:8888/anywhere-latest.tar.gz
On Server & Agent# mkdir /usr/local/anywhere && tar -xvf anywhere-latest.tar.gz -C /usr/local/anywhere


### Start Anywhere Server
On Server # cd /usr/local/anywhere
# Get Default Config
On Server # ./bin/anywhered config reset 
# Start Server
On Server # nohup ./bin/anywhered start &

### Start Anywhere Agent
On Agent >>  cd /usr/local/anywhere
On Agent >> nohup ./bin/anywhere -i anywhere-agent-1 -s 1.2.3.4 & # 1.2.3.4 shoud replace by your own Anywhere Server Public IP

### Add Tunnel Config
On Server >> cd /usr/local/anywhere
On Server >> ./bin/anywhered proxy add --agent-id anywhere-agent-1 --local-addr 10.0.0.2:80  --remote-addr 9497 --enable-wl false
# local-ip, local-port should replace by your pricate IP.

## Access Anywhere Server:9497, will be same as access 10.0.0.2:80
```

