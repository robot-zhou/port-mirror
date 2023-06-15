# port-mirror

#### 介绍
port-mirror 是一个通用的端口镜像服务， 可以将远端的服务器TCP端口，映射到本地TCP端口； 使用场景：

场景A:

Client -->  LocalServer(TCP Port) ---> TargetServer(TCP Port)

场景B:

Client -->  LocalServer(TCP Port) ---> Proxy(HTTP/Socks5) --> TargetServer(TCP Port)



#### 编译

go build

#### 使用说明

##### 启动命令
./port-mirror

##### 配置文件
默认路径: /etc/port-mirror.json
格式：JSON

```
{
    "mirror":[{
        "local":":<por>",               // 本的服务端口
        "target":"<hostname>:<port>",   // 远端服务机器名和端口
        "proxy":"<proxyserver1>,<proxyserver2>"     // 指定代理服务器，格式如： socks5://hostname:port, http://hostname:port
    }]
}
```