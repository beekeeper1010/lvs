# LVS - 本地视频播放服务

一个基于 **Go + Vue3** 的本地视频播放服务，前端打包产物直接嵌入 Go 二进制，实现**单文件部署**。

## 功能特性

- 内置账号登录（JWT 认证），无需注册/密码找回等用户管理
- 递归扫描本地目录的 MP4 文件，用 ffmpeg 自动提取缩略图
- 视频广场卡片式展示，分页浏览
- 基于 H5 `<video>` 标签播放，支持 Range 分片下载（拖拽进度、跳播流畅）
- 影院风格沉浸式播放页面
- 前端产物嵌入二进制，单文件即可部署运行

## 技术栈

| 模块 | 技术 |
| ---- | ---- |
| 后端 | Go + Gin + JWT + SQLite（`modernc.org/sqlite` 纯 Go 驱动，无 CGO） |
| 前端 | Vue3（组合式 API）+ Vite + vue-router + axios |

## 环境依赖

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+（仅构建前端时需要）
- [ffmpeg](https://ffmpeg.org/)（仅 `lvs scan` 提取缩略图时需要，运行服务不需要）

## 构建

```bash
# 1. 构建前端（产物输出到 web/dist）
cd web
npm install
npm run build
cd ..

# 2. 编译后端单文件
go build -o lvs        # Windows 下为 go build -o lvs.exe
```

## 使用

```bash
# 初始化：创建数据库并内置登录账号
./lvs init --username admin --password 123456
# 可选参数：--db <路径> 指定数据库文件位置（默认 ./lvs.db）

# 扫描视频：递归扫描目录及子目录的 mp4，提取缩略图入库
./lvs scan --dir /path/to/videos
# 可选参数：--db <路径>、--thumbs <缩略图目录>（默认 data/thumbs）

# 启动服务
./lvs            # 等价于 ./lvs serve
./lvs serve --port 8900    # 指定端口（默认 8900）
```

启动后访问 `http://localhost:8900`。

## API 接口

统一响应格式：`{"code": 0, "msg": "ok", "data": {...}}`，`code` 为 `0` 表示成功，其他为失败。

除登录接口外，均需在请求头携带 `Authorization: Bearer <token>`。

| 接口 | 方法 | 说明 |
| ---- | ---- | ---- |
| `/api/login` | POST | 登录，请求体 `{username, password}`，返回 JWT token |
| `/api/logout` | POST | 注销（前端删除本地 token） |
| `/api/video/list?page=1&pageSize=20` | GET | 分页获取视频列表，返回 `{list, total, page, pageSize}` |
| `/api/video/play?id=1&token=xxx` | GET | 播放视频流，支持 Range 分片（token 走 query 参数，因 `<video>` 标签无法携带 header） |
| `/api/video/thumb?id=1&token=xxx` | GET | 获取视频缩略图 |

## 开发模式（前端热更新）

```bash
# 终端 1：启动后端
./lvs serve

# 终端 2：启动前端 dev server（默认 5173 端口，支持局域网访问）
cd web && npm run dev
```

浏览器访问 `http://localhost:5173`，修改前端代码即时生效，`/api` 请求由 Vite 自动代理到后端（8900）。确认无误后需重新 `npm run build` + `go build` 生成正式单文件。

## 局域网访问

服务默认监听所有网卡，同网段设备访问 `http://<主机IP>:8900` 即可。注意放行 Windows 防火墙对应端口（8900 / 5173）。

## 目录结构

```text
lvs/
├── main.go         CLI 入口（init/scan/serve）
├── cmd.go          子命令参数解析
├── database.go     SQLite 建表、账号、JWT secret
├── auth.go         JWT 生成解析与认证中间件
├── handler.go      HTTP 接口
├── scan.go         MP4 递归扫描 + ffmpeg 缩略图
├── server.go       Gin 路由 + 前端产物嵌入
└── web/            Vue3 前端工程（构建产物 web/dist 被嵌入）
```

## 数据存储

- `lvs.db`：SQLite 数据库（视频记录、账号、JWT secret）
- `data/thumbs/`：缩略图文件（按视频路径 MD5 命名）
