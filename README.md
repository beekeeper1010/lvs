# lvs2

Local Video Service V2

## gops

### 安装

```sh
go get github.com/google/gops
# or
go install github.com/google/gops
```

### 使能

```go
opts := agent.Options{
   Addr:                   fmt.Sprintf(":%d", 9999),
   ShutdownCleanup:        true,
   ReuseSocketAddrAndPort: true,
  }
if err := agent.Listen(opts); err != nil {
log.Fatal(err)
}
```

### 使用

```sh
# 帮助
gops --help

# 查看所有go程序，有*标记的表示启用了调试
gops
# 12264 6220 main.exe * go1.19.2 D:\Download\tmp\go-build1334402605\b001\exe\main.exe
# 14404 5756 gops.exe   go1.19.2 C:\Users\2020\go\bin\gops.exe
# 6220  980  go.exe     go1.19.2 D:\Program Files\Go\bin\go.exe
# 10908 5568 gopls.exe  go1.19.2 C:\Users\2020\go\bin\gopls.exe

# 打印进程信息
gops 12264

# 内存分配和垃圾回收统计
gops memstats 12264
gops memstats 127.0.0.1:9999

# 运行时统计
gops stats 12264
gops stats 127.0.0.1:9999

# 打印调用栈
gops stack 12264
gops stack 127.0.0.1:9999

# 打印父子进程数
gops tree 12264
gops tree 127.0.0.1:9999
```

## MySQL

```sql
CREATE DATABASE ginbox CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci';
```

## 其他组件

+ 命令行用cobra
+ 配置文件用yaml(可以写注释，JSON不行)
+ 解析用viper
+ 结构化日志用lumberjack
+ 为应用程序安装windows服务用nssm(管理员模式) <http://nssm.cc/commands>
+ 二进制瘦身工具用upx

## 防火墙

```sh
firewall-cmd --add-port=50000-50100/udp --permanent
firewall-cmd --reload
firewall-cmd --list-all
```

## TODO

1. jwt
2. gorm √
3. websocket
4. redis
5. 自动化模块创建
6. 通过请求头中的timestamp和sign属性做简单鉴权 √

+ [ ] 登陆
+ [x] 命令行生成缩略图
+ [x] 命令行扫描mp4文件，写入数据库，不在配置文件中配置目录，避免频繁扫描
+ [ ] 前端按卡片展示，并显示缩略图
+ [ ] 前端点击缩略图弹框开始播放视频
+ [ ] 优化代码减少io
