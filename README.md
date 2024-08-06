# MaaResourceUtil

用网盘（天翼云）api下载[MaaResource](https://github.com/MaaAssistantArknights/MaaResource)

/maa/getResource的服务端请求路径现在是固定的，本地服务端测试访问：[http://127.0.0.1:8080/maa/getResource](http://127.0.0.1:8080/maa/getResource)


## 客户端配置文件

```yaml

directory: D:\Program Files\MAA-v5.3.1-win-x64 #这是maa的路径
getResourceUrls:
  - http://127.0.0.1:8080/maa/getResource #如果有服务端地址，这是服务端获取下载地址的路径
zipUrls:
  - https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip #一般情况下不需要修改，默认的maa resource github路径

```

## 服务端配置文件
```yaml
port: 8080 #服务端端口
zipUrl: https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip #一般情况下不需要修改，默认的maa resource github路径
cloud189:
  username: "" #天翼云盘用户名
  password: "" #天翼云盘密码

```


## 启动

①下载release的文件，windows使用cmd运行


②源码方式：

go mod vendor

服务端：[start_server.go](server%2Fcmd%2Fstart_server.go)

客户端：[start_client.go](client%2Fcmd%2Fstart_client.go)


## 需要注意的问题

切换网络环境时，appToken.json登录缓存不可复用。

rate limit默认不启用