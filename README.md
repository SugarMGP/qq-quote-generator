# QQ Quote Generator

一个使用 Go 编写的 QQ 引用图生成服务。服务直接计算消息布局、生成 SVG，再通过 resvg 输出 PNG；生产运行时不需要 Chromium，也没有浏览器进程池。

## 支持的功能

- `/png/` 返回 PNG 图片；
- `/base64/` 返回 PNG 的 Base64 文本；
- 支持字符串消息和图文消息段；
- 支持远程图片、data URI、自定义头像和 QQ 头像；
- 支持普通图片、emoji、sticker 的不同尺寸规则；
- 06:00–17:59 使用浅色主题，其余时间使用深色主题；
- 单张远程头像或消息图片加载失败时继续生成引用图；
- 字体按系统 family 解析，不在仓库内附带字体文件。

## 渲染结构

一次请求依次经过：

1. 解析并兼容原有消息 JSON；
2. 下载、校验并内嵌头像和消息图片；
3. 使用系统字体计算文本宽度和换行；
4. 计算卡片、消息行、头像、昵称、气泡和图片坐标；
5. 生成不引用外部资源的 SVG；
6. 通过 resvg 0.47.0 C API 渲染 RGBA，再编码为 PNG。

## 支持的平台

当前原生构建脚本和 CGO 链接参数支持：

- Windows amd64；
- Linux amd64；
- Docker Linux amd64。

resvg 静态库属于平台构建产物，生成在 `native/resvg/lib/<platform>/`，该目录不会提交到 Git。首次拉取代码后必须先构建原生库，才能执行 `go build`、`go run` 或视觉对比工具。

## Windows 本地构建

### 1. 安装依赖

需要以下工具，并确保命令能直接从 PowerShell 调用：

- Go 1.25 或更高版本；
- Rust 1.87 或更高版本；
- Cargo；
- Git；
- 支持 CGO 的 MinGW-w64 GCC。

检查环境：

```powershell
go version
rustc --version
cargo --version
git --version
gcc --version
```

### 2. 准备系统字体

服务按以下顺序选择第一个已经安装的字体：

1. `PingFang SC`
2. `Microsoft YaHei`
3. `Noto Sans CJK SC`

macOS 通常自带苹方，Windows 通常自带微软雅黑。三者都不存在时，服务会明确启动失败，不会换成不可控的字体。

### 3. 编译 resvg

```powershell
powershell -ExecutionPolicy Bypass -File native/resvg/build.ps1
```

脚本会执行以下工作：

- 从 resvg 官方仓库拉取固定标签 `v0.47.0`；
- 把源码缓存到 `%TEMP%\qq-quote-resvg-0.47.0`；
- 执行 `cargo build --release -p resvg-capi`；
- 生成 `native/resvg/lib/windows-amd64/libresvg.a`；
- 复制与该版本匹配的 `resvg.h`。

脚本不会自动选择其他 resvg 版本。编译失败时会直接停止。

### 4. 编译或启动服务

```powershell
go build -o qq-quote-go.exe .
./qq-quote-go.exe
```

开发时也可以直接运行：

```powershell
go run .
```

修改 resvg 版本、Rust 工具链或目标平台后，应重新执行 `build.ps1`。

## Linux 本地构建

### 1. 安装依赖

以 Alpine 为例：

```bash
apk add --no-cache go rust cargo git gcc musl-dev fontconfig font-noto-cjk
```

Debian/Ubuntu 可安装等价软件包：

```bash
sudo apt update
sudo apt install -y golang-go rustc cargo git gcc libc6-dev fontconfig fonts-noto-cjk
```

### 2. 编译 resvg

```bash
sh native/resvg/build.sh
```

脚本会固定拉取 `v0.47.0`，并生成：

```text
native/resvg/lib/linux-amd64/libresvg.a
```

### 3. 编译服务

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o qq-quote-go .
./qq-quote-go
```

如果启动时报“找不到可用字体”，请运行 `fc-list` 检查 `Noto Sans CJK SC` 是否已经安装。

## Docker 构建

Dockerfile 使用三个阶段：

1. Rust 阶段从官方 `v0.47.0` 编译 resvg C API；
2. Go 阶段启用 CGO，链接 resvg 静态库；
3. 运行阶段只保留服务、CA 证书、fontconfig 和 Noto CJK 字体。

构建镜像：

```bash
docker build -t qq-quote-go:resvg .
```

首次构建需要访问 Docker Hub、GitHub 和 crates.io。若在拉取基础镜像时出现 `failed to fetch anonymous token`，这是 Docker daemon 到 Docker Hub 的网络问题，并非 Go 或 resvg 编译错误。

启动容器：

```bash
docker run -d \
  --name qq-quote-go \
  --restart unless-stopped \
  -p 8080:5000 \
  qq-quote-go:resvg
```

查看日志：

```bash
docker logs -f qq-quote-go
```

## 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `5000` | HTTP 监听端口 |

## API

### POST `/png/`

返回 `image/png`：

```bash
curl -X POST http://localhost:8080/png/ \
  -H "Content-Type: application/json" \
  -d '[{"user_id":12345,"user_nickname":"张三","message":"Hello!"}]' \
  -o out.png
```

### POST `/base64/`

请求体与 `/png/` 相同，响应为纯 Base64 文本：

```bash
curl -X POST http://localhost:8080/base64/ \
  -H "Content-Type: application/json" \
  -d '[{"user_id":12345,"user_nickname":"张三","message":"Hello!"}]'
```

### 纯文本消息

```json
[
  {
    "user_id": 12345,
    "user_nickname": "张三",
    "message": "纯文本消息"
  }
]
```

### 图文混排

```json
[
  {
    "user_id": 12345,
    "user_nickname": "张三",
    "message": [
      {"type": "text", "text": "看这张图"},
      {"type": "image", "url": "https://example.com/image.jpg"},
      {"type": "image", "kind": "emoji", "url": "data:image/png;base64,..."},
      {"type": "image", "kind": "sticker", "url": "data:image/png;base64,..."}
    ]
  }
]
```

### 自定义头像

```json
[
  {
    "user_id": 0,
    "user_nickname": "匿名",
    "avatar": "data:image/png;base64,...",
    "message": "支持 URL 或 data URI 头像"
  }
]
```

`avatar` 为空且 `user_id > 0` 时，服务会请求 QQ 头像地址：

```text
https://q1.qlogo.cn/g?b=qq&nk=<user_id>&s=100
```

## 视觉回归工具

重构前的完整 Chromium 代码保存在 Git 提交 `49853a6`。对比工具会临时还原该提交，使用相同 JSON、相同本地图片和相同系统字体 family，分别请求旧版和新版服务。

```powershell
go run ./cmd/visual-regression \
  -fixture testdata/visual/messages.json \
  -out testdata/visual/out/current
```

输出文件：

- `chromium.png`：旧版输出；
- `resvg.png`：新版输出；
- `diff.png`：逐像素差异热图；
- `report.json`：尺寸和原始差异数据。

工具专用的旧版 Chromium 使用单 Page，避免本地首次预热多个 Page 时阻塞。Chromium 仅用于视觉回归，不会进入生产二进制或生产 Docker 镜像。
