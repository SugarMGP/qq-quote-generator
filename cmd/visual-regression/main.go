package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const legacyCommit = "49853a6"

var renderClient = &http.Client{Timeout: 30 * time.Second}

func main() {
	fixture := flag.String("fixture", "testdata/visual/messages.json", "request fixture")
	out := flag.String("out", "testdata/visual/out", "output directory")
	flag.Parse()
	root, err := os.Getwd()
	check(err)
	payload, err := loadFixture(root, *fixture)
	check(err)
	output, err := filepath.Abs(*out)
	check(err)
	check(os.MkdirAll(output, 0755))
	temp, err := os.MkdirTemp("", "qq-quote-legacy-")
	check(err)
	defer os.RemoveAll(temp)
	check(extractLegacy(root, temp))
	check(instrumentLegacy(temp))
	oldPort, err := freePort()
	check(err)
	oldPNG, err := renderWithService(temp, filepath.Join(temp, "legacy.exe"), oldPort, payload)
	check(err)
	newPort, err := freePort()
	check(err)
	newPNG, err := renderWithService(root, filepath.Join(temp, "resvg.exe"), newPort, payload)
	check(err)
	check(os.WriteFile(filepath.Join(output, "chromium.png"), oldPNG, 0644))
	check(os.WriteFile(filepath.Join(output, "resvg.png"), newPNG, 0644))
	oldImage, _, err := image.Decode(bytes.NewReader(oldPNG))
	check(err)
	newImage, _, err := image.Decode(bytes.NewReader(newPNG))
	check(err)
	report, diff := Compare(oldImage, newImage)
	check(writePNG(filepath.Join(output, "diff.png"), diff))
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	check(err)
	check(os.WriteFile(filepath.Join(output, "report.json"), reportJSON, 0644))
	fmt.Println(string(reportJSON))
	if !report.SameSize {
		os.Exit(2)
	}
}

func loadFixture(root, path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var messages []map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}
	base := filepath.Dir(path)
	for _, message := range messages {
		if avatar, ok := message["avatar"].(string); ok {
			message["avatar"], err = localDataURI(base, avatar)
			if err != nil {
				return nil, err
			}
		}
		if segments, ok := message["message"].([]interface{}); ok {
			for _, raw := range segments {
				if segment, ok := raw.(map[string]interface{}); ok {
					if source, ok := segment["url"].(string); ok && source != "missing.png" {
						segment["url"], err = localDataURI(base, source)
						if err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}
	return json.Marshal(messages)
}

func localDataURI(base, source string) (string, error) {
	if strings.HasPrefix(source, "data:") || strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return source, nil
	}
	data, err := os.ReadFile(filepath.Join(base, source))
	if err != nil {
		return "", err
	}
	return "data:" + http.DetectContentType(data) + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func extractLegacy(root, target string) error {
	archive := filepath.Join(target, "legacy.zip")
	command := exec.Command("git", "archive", "--format=zip", "-o", archive, legacyCommit)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("archive legacy: %w: %s", err, output)
	}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		path := filepath.Join(target, filepath.FromSlash(file.Name))
		if !strings.HasPrefix(path, target+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe archive path %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		destination, err := os.Create(path)
		if err != nil {
			source.Close()
			return err
		}
		_, copyErr := io.Copy(destination, source)
		source.Close()
		destination.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func instrumentLegacy(target string) error {
	path := filepath.Join(target, "renderer.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.Replace(string(data), `"html/template"`, "\"html/template\"\n\t\"log\"", 1)
	replacements := []struct{ old, new string }{
		{`processed, err := r.processMessages(messages)`, `log.Print("legacy: process messages"); processed, err := r.processMessages(messages)`},
		{`page, err := r.pool.Acquire(ctx)`, `log.Print("legacy: acquire page"); page, err := r.pool.Acquire(ctx); log.Print("legacy: page acquired")`},
		{`if err := renderPage.Navigate("about:blank"); err != nil {`, `log.Print("legacy: navigate"); if err := renderPage.Navigate("about:blank"); err != nil {`},
		{`if err := renderPage.SetDocumentContent(html); err != nil {`, `log.Print("legacy: set content"); if err := renderPage.SetDocumentContent(html); err != nil {`},
		{`_ = renderPage.WaitIdle(500 * time.Millisecond)`, `_ = renderPage.WaitIdle(500 * time.Millisecond); log.Print("legacy: idle")`},
		{`png, err = el.Screenshot`, `log.Print("legacy: screenshot"); png, err = el.Screenshot`},
	}
	for _, replacement := range replacements {
		text = strings.Replace(text, replacement.old, replacement.new, 1)
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func renderWithService(workdir, executable, port string, payload []byte) ([]byte, error) {
	build := exec.Command("go", "build", "-o", executable, ".")
	build.Dir = workdir
	if output, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("build %s: %w: %s", workdir, err, output)
	}
	command := exec.Command(executable)
	command.Dir = workdir
	command.Env = append(os.Environ(), "PORT="+port, "POOL_SIZE=1")
	var logs bytes.Buffer
	command.Stdout, command.Stderr = &logs, &logs
	if err := command.Start(); err != nil {
		return nil, err
	}
	defer stopProcess(command)
	url := "http://127.0.0.1:" + port
	deadline := time.Now().Add(90 * time.Second)
	for {
		response, err := http.Get(url)
		if err == nil {
			response.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("service start timeout: %s", logs.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
	response, err := renderClient.Post(url+"/png/", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("render request: %w; logs: %s", err, logs.String())
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("render HTTP %d: %s; logs: %s", response.StatusCode, data, logs.String())
	}
	return data, nil
}

func freePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	return fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port), nil
}

func stopProcess(command *exec.Cmd) {
	if command.Process == nil {
		return
	}
	if strings.EqualFold(filepath.Ext(command.Path), ".exe") {
		_ = exec.Command("taskkill", "/PID", fmt.Sprint(command.Process.Pid), "/T", "/F").Run()
	} else {
		_ = command.Process.Kill()
	}
	_, _ = command.Process.Wait()
}

func writePNG(path string, image image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, image)
}
func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
