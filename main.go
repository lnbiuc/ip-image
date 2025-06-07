package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

// IPInfo 定义从 ip-api.com 获取的 JSON 响应结构
type IPInfo struct {
	Query      string `json:"query"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	Org        string `json:"org"`
	AS         string `json:"as"`
	Status     string `json:"status"`
	Message    string `json:"message"` // 错误信息
}

// 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	//return "222.90.15.11"
	// 优先检查 X-Forwarded-For 或 X-Real-IP（代理环境下常用）
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		return xrip
	}

	// fallback: RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// 查询 IP 定位信息
func getIPInfo(ip string) (*IPInfo, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Status != "success" {
		return nil, fmt.Errorf("查询失败: %s", info.Message)
	}
	return &info, nil
}

type pngBufferWriter struct {
	b *strings.Builder
}

func (w *pngBufferWriter) Write(p []byte) (n int, err error) {
	return w.b.WriteString(string(p))
}

// HTTP 处理函数
func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("==收到请求==")
	log.Println("正在获取IP地址")
	ip := getClientIP(r)

	info, err := getIPInfo(ip)
	if err != nil {
		http.Error(w, "IP信息获取失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("获取到IP信息: " + ip)

	//timestamp := time.Now().Format("2006-01-02 15:04:05")
	now := time.Now()

	// 获取中文星期
	weekdays := map[time.Weekday]string{
		time.Sunday:    "星期日",
		time.Monday:    "星期一",
		time.Tuesday:   "星期二",
		time.Wednesday: "星期三",
		time.Thursday:  "星期四",
		time.Friday:    "星期五",
		time.Saturday:  "星期六",
	}

	timestamp := now.Format("2006年01月02日") + " " + weekdays[now.Weekday()]

	const width = 350
	const height = 120
	log.Println("开始生成图片")
	dc := gg.NewContext(width, height)

	dc.SetRGB(1, 1, 1) // 白色背景
	dc.Clear()

	dc.SetRGB(0, 0, 0)
	fontPath := "./MapleMono-NF-CN-Regular.ttf"
	if err := dc.LoadFontFace(fontPath, 15); err != nil {
		http.Error(w, "字体加载失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lines := []struct {
		label string
		value string
	}{
		{"今天是: ", timestamp},
		{"您的地址是: ", fmt.Sprintf("%s %s %s", info.Country, info.RegionName, info.City)},
		{"您的IP是: ", info.Query},
		{"运营商信息: ", info.Org},
		{"AS信息: ", info.AS},
	}

	y := 25.0
	for _, line := range lines {
		dc.SetRGB(0, 0, 0) // 黑色文字
		dc.DrawString(line.label, 10, y)

		// 获取 label 的宽度，用于计算 value 的起始位置
		labelWidth, _ := dc.MeasureString(line.label)

		dc.SetRGB(1, 0, 0) // 红色文字
		dc.DrawString(line.value, 10+labelWidth, y)

		y += 20
	}
	log.Println("图片生成完成")
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, dc.Image()); err != nil {
		http.Error(w, "图像生成失败", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/AQuVk4853gdCarY6ysbscNZCL4A7ndgK", handler)
	fmt.Println("Server running at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
