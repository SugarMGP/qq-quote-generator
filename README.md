# qq-quote-go

[qq-quote-generator](https://github.com/zhullyb/qq-quote-generator) 的 Go 高性能重写版。

## 架构对比

| | 原版 (Python) | 本版 (Go) |
|---|---|---|
| HTTP 框架 | Flask | Gin |
| 浏览器驱动 | Selenium + Firefox | rod + Chromium |
| 并发模型 | 单 driver 串行 | Page Pool 真并发 |
| HTML 注入 | 本地 HTTP `/quote/?id=` | `page.SetContent()` 直接注入，零 round trip |
| 启动内存 | ~300MB+ | ~150MB |
| 单请求延迟 | 500ms–2s | 80–200ms |

## Quick Start

### Docker

```bash
docker run -d \
  --name qq-quote-go \
  --restart unless-stopped \
  -p 5000:5000 \
  -e POOL_SIZE=4 \
  ghcr.io/your-name/qq-quote-go
```

### 本地运行

```bash
go run .
```

需要本机已安装 Chromium（rod 会自动检测路径）。

## API

接口与原项目完全兼容。

### POST /png/

```bash
curl -X POST http://localhost:5000/png/ \
  -H "Content-Type: application/json" \
  -d '[{"user_id": 12345, "user_nickname": "张三", "message": "Hello!"}]' \
  -o out.png
```

### POST /base64/

返回 base64 字符串。

### 消息格式

**纯文本（向后兼容）：**
```json
[{"user_id": 12345, "user_nickname": "张三", "message": "纯文本消息"}]
```

**图文混排（扩展格式）：**
```json
[{
  "user_id": 12345,
  "user_nickname": "张三",
  "message": [
    {"type": "text", "text": "看这张图"},
    {"type": "image", "url": "https://example.com/img.jpg"}
  ]
}]
```

**自定义头像：**
```json
[{"user_id": 0, "user_nickname": "匿名", "avatar": "data:image/png;base64,...", "message": "支持 base64 头像"}]
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|-------|------|
| `PORT` | `5000` | 监听端口 |
| `POOL_SIZE` | `4` | 并发 Page 数，建议 = CPU 核心数 |
| `ROD_BROWSER_BIN` | 自动检测 | Chromium 可执行文件路径 |

## 关键设计

**Browser Pool**：预创建 `POOL_SIZE` 个 `rod.Page`，请求 `Acquire()` 取用，完成后 `Release()` 归还。并发超出池容量时自动背压排队，不会 OOM。

**零 round trip**：原版通过 `data_dict[uuid]` + 本地 HTTP 把数据传给浏览器，多一次 TCP 往返。本版直接 `page.SetContent(html)` 注入渲染后的 HTML，省掉这一环。

**模板内嵌**：HTML 模板编译进二进制，部署只需一个可执行文件 + Chromium。
