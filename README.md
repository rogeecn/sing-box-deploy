## sing-box + Caddy 一键部署脚本

本仓库提供一个可定制域名的 sing-box + Caddy 部署示例，并在部署完成后自动生成常见协议 (VMess/VLESS) 的订阅链接。

### 环境准备

```
# install caddy

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg
chmod o+r /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# install sing-box

curl -fsSL https://sing-box.app/install.sh | sh
```

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
go run . deploy your.domain.com --email admin@your.domain.com
```

主要子命令：

- `deploy <domain>`：渲染 sing-box 入站、`config.json`、Caddyfile 以及订阅文件；若 `<root>/tls.key|tls.cer` 缺失，会自动执行 `sing-box generate tls-keypair <domain> -m 1024` 生成自签证书，并在模板中引用实际路径，同时为每个入站随机分配高位端口。命令会先列出所有支持的协议，输入编号即可部署任意组合（留空等同于全部），部署完成后会把所选协议的分享链接直接打印出来。常用参数：
  - `--type` (可重复)：指定入站类型，默认全部 (如 `vless-ws-tls`、`vmess-h2-tls` 等)。
  - `--name`：订阅展示名称 (默认 `<domain>`)。
  - `--root`：sing-box 目录 (默认 `/etc/sing-box`)。
  - `--caddy`：Caddyfile 输出路径 (默认 `/etc/caddy/Caddyfile`)。
  - `--subscriptions`：订阅文件目录 (默认 `/etc/sing-box/subscriptions`)。
  - `--sing-box-bin`：`sing-box` 二进制路径 (默认查找 PATH)。
- `list`：读取状态文件，列出已部署的入站、监听端口及路径。
- `url`：打印订阅链接，同时输出一个在线二维码图片地址 (基于 `api.qrserver.com`)。

CLI 会把部署记录保存到 `--state` 指定的 JSON 文件 (默认 `sing-box-state.json`)，`list` 与 `url` 子命令据此展示数据。

部署完成后会生成：

- `sing-box` 主配置：`<root>/00_common.json`（仅保留日志/出站/路由），入站碎片以 `02_inbounds_*.json` 命名直接放在 `<root>/` 下，每个文件都是 `{"inbounds": [...]}` 结构，可直接被 `sing-box -C` 自动加载；
- `Caddyfile`：`--caddy` 指定位置；
- 订阅链接：`--subscriptions` 目录中的 `<domain>.txt`；`url` 子命令也会将每条链接对应的二维码 URL 打印出来。

运行服务时可使用 `sing-box -C <root> run`，sing-box 会自动加载 `<root>` 目录下所有配置文件。

### 常见问题

- **重复执行脚本是否安全？**
  会自动备份现有的 `/etc/caddy/Caddyfile` 与 `/etc/sing-box/config.json`，以 `.bak.<timestamp>` 形式保存，方便回滚。

- **如何查看订阅链接？**
  执行 `sudo cat /etc/sing-box/subscriptions/<domain>.txt` 即可，里面包含每个协议的分享 URL。

- **如何更新 sing-box？**
  再次运行脚本并指定 `--sing-box-version`，脚本会下载对应版本并覆盖旧二进制，然后重新渲染配置并重启服务。

如需扩展更多协议，可在 `configs/` 目录新增模板 JSON (包含占位域名 `v9.20140202.xyz`)，脚本会自动检测并生成对应配置与链接。

### CI/CD

仓库包含 `.github/workflows/release.yml`，在 `main` 分支有新的 push 时会自动执行 `go test`, `go build`，并以提交哈希为 tag (`auto-<sha>`) 发布预发行版，附带打包后的 `sing-box-deploy` 可执行文件。无需手动干预即可获取最新二进制。
