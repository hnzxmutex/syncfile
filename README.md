# 简单文件同步工具

同时在服务器和本地执行该工具,可以同步本地目录到服务器目录

-----
## 功能列表

> * 目录同步
> * 简单加密
> * 密码校验
> * 黑名单过滤
> * 支持mac/win/linux

## 参数说明
$ ./bin/syncfile -h
Usage of ./bin/syncfile:
  -d	server mode 可选 设置了表示为服务器模式
  -dir string 必选 同步目录
    	server or client sync fold (default "./gosync/")
  -host string 服务器ip,客户端模式需要配置
    	server host
  -i string 黑名单列表,格式为正则,参见ignore.ini
    	ignore file (default "./ignore.ini")
  -p string 密码,服务端和客户端需一致
    	password (default "tgideas")
  -port string 服务器端口
    	server listen port (default "443")


## 使用

已经预编译在bin目录
可以重新开发编译
win:make windows
mac:make
linux:make linux

## 例子

server:
```
./bin/syncfile -d -dir /tmp/server -port 4433
```


client:
```
./bin/syncfile -host 127.0.0.1 -port 4433 -dir /tmp/client -i /tmp/ignore.ini
```
