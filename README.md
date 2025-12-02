## sing-box + Caddy 一键部署脚本

本仓库提供一个可定制域名的 sing-box + Caddy 部署示例，并在部署完成后自动生成常见协议 (VMess/VLESS) 的订阅链接。

### 功能概览

- 使用内置模板渲染 sing-box 入站配置、Caddyfile 以及订阅链接；
- 自动安装 Caddy、sing-box 及依赖，创建 systemd 服务；
- 针对每个协议生成唯一的路径/UUID，并将订阅链接保存到 `/etc/sing-box/subscriptions/<domain>.txt`；
- 部署完毕后自动启动/重载相关服务。

### 使用前准备

1. 服务器需为 Debian/Ubuntu 系 (使用 `apt`) 且具有公网 IPv4；
2. 将目标域名的 A 记录解析到服务器；
3. 使用 `root` 用户或具备 `sudo` 权限；
4. 确保 80/443 端口未被其他程序占用。

### CLI 快速开始

项目内置基于 Cobra 的 CLI，可用于一键渲染配置、查看已部署入站和生成订阅链接/二维码 URL。编译或直接运行：

```bash
git clone https://example.com/sing-box-deploy.git
cd sing-box-deploy
go run . deploy --domain your.domain.com --email admin@your.domain.com
```

主要子命令：

- `deploy`：渲染 sing-box 入站、`config.json`、Caddyfile 以及订阅文件；若 `<root>/tls.key|tls.cer` 缺失，会自动执行 `sing-box generate tls-keypair <domain> -m 1024` 生成自签证书，并在模板中引用实际路径，同时为每个入站随机分配高位端口。常用参数：
  - `--domain/-d` (必填)：目标域名。
  - `--type` (可重复)：指定入站类型，默认全部 (如 `vless-ws-tls`、`vmess-h2-tls` 等)。
  - `--root`：sing-box 目录 (默认 `/etc/sing-box`)。
  - `--caddy`：Caddyfile 输出路径 (默认 `/etc/caddy/Caddyfile`)。
  - `--subscriptions`：订阅文件目录 (默认 `/etc/sing-box/subscriptions`)。
  - `--sing-box-bin`：`sing-box` 二进制路径 (默认查找 PATH)。
- `list`：读取状态文件，列出已部署的入站、监听端口及路径。
- `url`：打印订阅链接，同时输出一个在线二维码图片地址 (基于 `api.qrserver.com`)。

CLI 会把部署记录保存到 `--state` 指定的 JSON 文件 (默认 `sing-box-state.json`)，`list` 与 `url` 子命令据此展示数据。

部署完成后会生成：

- `sing-box` 主配置：`<root>/config.json`（仅保留日志/出站/路由），入站碎片位于 `<root>/configs/`；
- `Caddyfile`：`--caddy` 指定位置；
- 订阅链接：`--subscriptions` 目录中的 `<domain>.txt`；`url` 子命令也会将每条链接对应的二维码 URL 打印出来。

运行服务时可使用 `sing-box -C <root>/configs run -c <root>/config.json`，sing-box 会自动加载 `<root>/configs` 目录下的全部入站配置文件。

### 常见问题

- **重复执行脚本是否安全？**  
  会自动备份现有的 `/etc/caddy/Caddyfile` 与 `/etc/sing-box/config.json`，以 `.bak.<timestamp>` 形式保存，方便回滚。

- **如何查看订阅链接？**  
  执行 `sudo cat /etc/sing-box/subscriptions/<domain>.txt` 即可，里面包含每个协议的分享 URL。

- **如何更新 sing-box？**  
  再次运行脚本并指定 `--sing-box-version`，脚本会下载对应版本并覆盖旧二进制，然后重新渲染配置并重启服务。

如需扩展更多协议，可在 `configs/` 目录新增模板 JSON (包含占位域名 `v9.20140202.xyz`)，脚本会自动检测并生成对应配置与链接。
