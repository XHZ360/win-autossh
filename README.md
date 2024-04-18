# Win-AutoSSH
> A tool to forward tcp port
## Usage
### config file
> Configure before run the tool, the config file named `config.yaml` in the same directory with the tool.
```yaml
#local-address:local-port,remote-address:remote-port,
mappings:
  # remote host listening on remote-address:remote-port and forward to [local-address:local-port] on local machine
  # remote host sshd.config should set GatewayPorts `clientspecified` otherwise the listening address will not be specified
  rtl:
    - "localhost:5173,0.0.0.0:5173"
  # local machine listening on local-address:local-port and forward to remote-address:remote-port on server
  ltr:
    - "0.0.0.0:5173,localhost:5173"
server:
  addr: "hostname:port"
  user: "proxy"
  password: "*****"
  keyfile: "C:/Users/yourAccount/.ssh/id_rsa"
```

### command line
``` bash
win-autossh.exe 
```
### run as service
> logfiles located in `./logs` in the same directory with the tool;\
> install / uninstall service should with administrator privilege
```bash
# install as service
win-autossh.exe -s install
# uninstall service
win-autossh.exe -s uninstall
# start service
win-autossh.exe -s start
# stop service
win-autossh.exe -s stop
```