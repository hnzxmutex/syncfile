# 简单文件同步工具

同时在服务器和本地执行该工具,可以同步本地目录到服务器目录

-----
## 功能列表

> * 目录同步
> * 多目录同步
> * 简单加密
> * 密码校验
> * 黑名单过滤
> * 支持mac/win/linux

### Usage:

  syncfile server [flags] 服务端模式,接受客户端上传的文件
  syncfile client [flags] 客户端模式,监控本地目录和文件改动并上传

### 服务端模式Flags:

> *  -c, --config string   服务端配置文件
> *  -h, --help            help for server

### 客户端模式Flags:

> *         --debug             是否打印debug信息
> *     -d, --dir string        同步目录
> *     -h, --help              help for client
> *         --host string       服务器IP
> *     -i, --ignore string     忽略上传的文件列表,内容为正则 (default "./ignore.ini")
> *         --password string   同步口令 (default "tgideas")
> *     -p, --port string       服务器端口 (default "443")
> *     -w, --watch             是否持续监控文件改动并同步 *   -info bool 可选，默认false，显示debug信息


## 使用

已经预编译在bin目录
可以重新开发编译
win:make windows
mac:make mac
linux:make linux

## 例子

### server:

```
./bin/syncfile server -c ./server.yaml
```


### client:

```
./bin/syncfile client --host 127.0.0.1 -d ./sync_dir -i ./syncignore.ini -w -p 8081 --password app_foo_password
```

### 服务端配置,yaml格式

```yaml
port: 8081 #listen端口

app_list: #有效的配置项,必填
    - app_foo
    - app_bar #跟下面的名字匹配

app_foo: #配置项1
    password: app_foo_password
    path: /tmp/b/
    #ignore_config_file: /tmp/ignore.ini

app_bar: #配置项2
    password: xcnsdi!3431
    path: /tmp/a/
```
