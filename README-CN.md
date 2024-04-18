# Win-AutoSSH
Win-AutoSSH 是一款用于转发 TCP 端口的工具。

## 使用说明
### 配置文件
在运行工具之前，请先进行配置。配置文件 config.yaml 位于工具的同一目录下。

```yaml
# 格式：本地地址:本地端口, 远程地址:远程端口
mappings:
# 远程主机在 remote-address:remote-port 上监听，并将其转发到本地机器的 local-address:local-port
# 远程主机的 sshd.config 应设置 GatewayPorts 为 `clientspecified`，否则监听地址将不会指定
rtl:
- "localhost:5173,0.0.0.0:5173"
# 本地机器在 local-address:local-port 上监听，并将其转发到远程的 remote-address:remote-port
ltr:
- "0.0.0.0:5173,localhost:5173"
server:
addr: "主机名:端口"
user: "代理用户"
password: "*****"
keyfile: "C:/Users/您的账户名/.ssh/id_rsa"
```
### 命令行参数
运行 Win-AutoSSH 的命令行如下：
```bash
win-autossh.exe
```
### 作为服务运行
日志文件位于与工具同一目录下的 ./logs 文件夹中；安装/卸载服务需要管理员权限。

```bash
# 安装为服务
win-autossh.exe -s install
# 卸载服务
win-autossh.exe -s uninstall
# 启动服务
win-autossh.exe -s start
# 停止服务
win-autossh.exe -s stop
```
注意：在执行服务相关的命令时，请确保您具有管理员权限。