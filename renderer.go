package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

const defaultRenderTimeout = 8 * time.Second

// Renderer 持有 HTML 模板和 Page 池的引用
type Renderer struct {
	tmpl *template.Template
	pool *BrowserPool
}

func NewRenderer(pool *BrowserPool) (*Renderer, error) {
	tmpl, err := template.New("quote").Parse(quoteHTML)
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl, pool: pool}, nil
}

// Render 处理一批消息，返回 PNG bytes
func (r *Renderer) Render(ctx context.Context, messages []Message) (png []byte, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRenderTimeout)
	defer cancel()

	// 1. 预处理消息（解析 message 字段、拼装头像 URL）
	processed, err := r.processMessages(messages)
	if err != nil {
		return nil, err
	}

	// 2. 渲染 HTML 模板到字符串
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, renderData{Messages: processed, Theme: themeForTime(time.Now()).Class}); err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	html := buf.String()

	// 3. 从池中取一个 Page，注入 HTML，截图，归还
	page, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire page: %w", err)
	}
	pageOK := false
	defer func() {
		if pageOK {
			r.pool.Release(page)
			return
		}
		r.pool.Replace(page)
	}()

	renderPage := page.Context(ctx)

	// SetContent 直接注入 HTML，完全避免本地 HTTP round trip
	// rod 的 Navigate + SetDocumentContent 方式
	if err := renderPage.Navigate("about:blank"); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	// 用 CDP 直接设置页面内容
	if err := renderPage.SetDocumentContent(html); err != nil {
		return nil, fmt.Errorf("setContent: %w", err)
	}

	// 等待页面短暂空闲；外部图片慢或失效时不能阻塞整个请求。
	_ = renderPage.WaitIdle(500 * time.Millisecond)

	// 只截取 #app 元素，高度自适应内容
	el, err := renderPage.Element("#app")
	if err != nil {
		return nil, fmt.Errorf("element #app: %w", err)
	}
	png, err = el.Screenshot(proto.PageCaptureScreenshotFormatPng, 90)
	if err != nil {
		return nil, fmt.Errorf("screenshot: %w", err)
	}

	pageOK = true
	return png, nil
}

// RenderBase64 返回 base64 编码的 PNG
func (r *Renderer) RenderBase64(ctx context.Context, messages []Message) (string, error) {
	png, err := r.Render(ctx, messages)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(png), nil
}

// processMessages 将原始 Message 列表转换为模板可用的结构
func (r *Renderer) processMessages(messages []Message) ([]processedMessage, error) {
	result := make([]processedMessage, 0, len(messages))

	for _, msg := range messages {
		pm := processedMessage{
			Nickname: msg.UserNickname,
			Avatar:   safeImageURL(resolveAvatar(msg)),
		}

		segs, err := parseMessageField(msg.Message)
		if err != nil {
			return nil, err
		}
		pm.Segments = processMessageSegments(segs)

		result = append(result, pm)
	}
	return result, nil
}
