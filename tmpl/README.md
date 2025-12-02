# Template Layout

`tmpl/` 目录保存 Go `text/template` 模板，用于生成 sing-box 配置片段和 Caddyfile。

## 占位变量

渲染器会向模板提供如下数据结构：

- `.Domain` (`string`): 目标域名，例如 `example.com`。
- `.Email` (`string`): 申请证书所用邮箱，可为空。
- `.Inbounds` (`map[string]InboundSpec`): 不同协议入站的规格，键对应模板文件名去掉后缀，例如 `vless-ws-tls`。

`InboundSpec` 结构：

```go
 type InboundSpec struct {
     Tag        string // VLESS-WS-TLS-example.com
     Listen     string // 默认 127.0.0.1
     ListenPort int
     UUID       string
     Password   string // trojan/shadowsocks 等可选字段
     Path       string // 以 / 开头
     Host       string // 默认与 Domain 相同
     Transport  string // ws/http/httpupgrade
 }
```

模板中可使用 `{{ with index .Inbounds "vless-ws-tls" }}` 获取对应协议的具体数值。没有使用的协议可以在渲染时从 `.Inbounds` 中省略。

## 目录结构

```
tmpl/
├── README.md
├── caddy/
│   └── site.caddy.tmpl        # 生成 Caddyfile
└── sing-box/
    └── inbounds/
        ├── vmess-h2-tls.json.tmpl
        ├── vmess-httpupgrade-tls.json.tmpl
        ├── vmess-ws-tls.json.tmpl
        ├── vless-h2-tls.json.tmpl
        ├── vless-httpupgrade-tls.json.tmpl
        └── vless-ws-tls.json.tmpl
```

后续需要支持新的协议时，在 `sing-box/inbounds` 下新增 `.json.tmpl` 文件即可。
