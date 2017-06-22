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

> *   -d bool server mode 可选,设置了表示为服务器模式
> *   -dir string 必选,同步目录,默认./gosync/
> *   -host string 服务器ip,客户端模式需要配置
> *   -i string 黑名单列表,格式为正则,参见ignore.ini,默认为./ignore.ini
> *   -p string 密码,服务端和客户端需一致，默认为tgideas
> *   -port string 服务器端口,默认443
> *   -w bool 可选，默认false，客户端持续监听目录变化并同步
> *   -info bool 可选，默认false，显示debug信息


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
