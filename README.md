# MaaResourceUtil

用网盘（天翼云）api下载[MaaResource](https://github.com/MaaAssistantArknights/MaaResource)

## 客户端配置文件
```yaml

directory: D:\Program Files\MAA-v5.3.1-win-x64 #这是maa的路径
getResourceUrls:
  - http://127.0.0.1:8080/maa/getResource #如果有服务端地址，这是服务端获取下载地址的路径
zipUrls:
  - https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip #默认的maa resource github路径

```

## 服务端配置文件
```yaml
port: 8080 #服务端端口
zipUrl: https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip #默认的maa resource github路径
cloud189:
  username: "" #天翼云盘用户名
  password: "" #天翼云盘密码

```


## 启动

#### 服务端：[start_server.go](server%2Fcmd%2Fstart_server.go)

#### 客户端：[start_client.go](client%2Fcmd%2Fstart_client.go)