# ngrok
映射内网的端口到外网，复用单通道tcp，配合nginx可以转发http请求到内网服务器做调试用 如微信服务器调试
由于本程序只是tcp的通道，所以并没有处理http协议请求头host, 可以配合nginx转发来实现相同的效果 proxy_pass


# HTTP服务支持
`目前增加了HTTP转发支持` 启用请参考`main.go`里面的注释，重写`fasthttp.client`的`dial`方法即可

# 配置

配置config.json

- Mode : 模式 运行在服务端或者客户端  server or client
- Smux_addr : 服务端通讯用的地址
- Listen_addr : 服务端处理请求的地址
- Local_addr : 客户端转发到本地的地址

# 使用

```
    配置好GOPATH环境变量
    go get
	go build
	nohup ./ngrok -config config.json &
```
