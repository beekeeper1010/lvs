# lvs2

`lvs2`是`Local Video Service`的`v2`版本，用于本地`mp4`文件的托管

## 前提

1. `mp4`文件视频编码格式为`h264`，音频编解格式为`aac`
2. 对于非`mp4`文件或编码格式不满足上述要求的情况，可以使用[ffmpeg](https://github.com/BtbN/FFmpeg-Builds/releases)工具转换

## 特性

1. 扫描用户指定目录下的`mp4`文件，生成`sqlite3`数据库档案(文件名、路径、大小、播放时长、缩略图)
2. 基于`gin`提供播放服务，通过解析请求头中的`Range`属性实现分片下载播放

## 构建

```bash
git clone https://github.com/beekeeper1010/lvs2.git
cd lvs2
go build -ldflags="-s -w"
```

## 使用

+ 扫描mp4文件

  ```bash
  # 扫描目录1和目录2中的mp4文件，过滤掉小于60秒的视频，指定缩略图高度为100px，生成lvs2.db数据库档案
  lvs2 scan --dir=目录1 --dir=目录2 --filter=60 --height=100 --db=lvs2.db
  ```

+ 运行服务

  ```bash
  # 基于lvs2.db数据库档案启动视频服务，监听8080端口，并将日志输出到lvs2.log文件中
  lvs2 run --addr=:8080 --db=lvs2.db --log=lvs2.log
  ```

+ 自动补全

  ```powershell
  # 以powershell为例，使用lvs2命令自动补全
  lvs2 completion powershell | Out-String | Invoke-Expression
  ```

+ 通过`lvs2 -h`和`lvs2 help command`来查看完整的帮助信息

  ```text
  lvs2 is a Local Video Service

  Usage:
  lvs2 [command]

  Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Run server
  scan        Scan mp4 files to generate sqlite db file

  Flags:
  -h, --help      help for lvs2
  -v, --version   version for lvs2

  Use "lvs2 [command] --help" for more information about a command.
  ```

## 组件依赖

| 组件 | 必选 | 说明 |
| --- | --- | --- |
| [ffmpeg](https://github.com/BtbN/FFmpeg-Builds/releases) | Y | 用于获取播放时长、生成缩略图，测试版本`4.4.x` |
| [nssm](https://nssm.cc/download) | N | 用于安装Windows服务，测试版本`2.24` |
| [upx](https://github.com/upx/upx/releases/) | N | 用于二进制瘦身，测试版本`4.2.2` |

## TODO

### 后端

+ [ ] 鉴权
+ [ ] IO性能优化
+ [ ] 弹幕
+ [ ] 其他

### 前端

+ [ ] 用户管理
+ [ ] 播放页面，卡片中展示缩略图、文件名等
+ [ ] 搜索
+ [ ] 适配移动设备
