package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

// ip-api.com 返回结构
type IPInfo struct {
	Query      string `json:"query"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	Org        string `json:"org"`
	AS         string `json:"as"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// 百度API返回结构
type BaiduAPIResponse struct {
	Status  int    `json:"status"`
	Address string `json:"address"`
	Content struct {
		Address string `json:"address"`
	} `json:"content"`
}

var baiduAK string // 百度API key

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		return xrip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

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
	log.Println(info)
	if info.Status != "success" {
		return nil, fmt.Errorf("ip-api 查询失败: %s", info.Message)
	}
	return &info, nil
}

func getBaiduAddress(ip, ak string) (string, error) {
	url := fmt.Sprintf("https://api.map.baidu.com/location/ip?ip=%s&coor=bd09ll&ak=%s", ip, ak)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result BaiduAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	log.Println(result)
	if result.Status != 0 {
		return "", fmt.Errorf("百度API返回状态错误: %d", result.Status)
	}
	return result.Content.Address, nil
}

func renderImage(w http.ResponseWriter, ip string, info *IPInfo, address string) {
	now := time.Now()
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

	const width = 550
	const height = 120
	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1)
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
		{"您的地址是: ", address},
		{"您的IP是: ", info.Query},
		{"运营商信息: ", info.Org},
		{"AS信息: ", info.AS},
	}

	y := 25.0
	for _, line := range lines {
		dc.SetRGB(0, 0, 0)
		dc.DrawString(line.label, 10, y)
		labelWidth, _ := dc.MeasureString(line.label)
		dc.SetRGB(1, 0, 0)
		dc.DrawString(line.value, 10+labelWidth, y)
		y += 20
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, dc.Image()); err != nil {
		http.Error(w, "图像生成失败", http.StatusInternalServerError)
	}
}

// 支持自动识别客户端IP
func handler(w http.ResponseWriter, r *http.Request) {
	ip := getClientIP(r)
	serveWithIP(w, ip)
}

// 支持通过参数 ?ip=xxx 指定IP
func handlerWithQuery(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "请通过 ?ip= 参数提供 IP 地址", http.StatusBadRequest)
		return
	}
	serveWithIP(w, ip)
}

// 核心处理逻辑
func serveWithIP(w http.ResponseWriter, ip string) {
	log.Println("处理IP:", ip)

	info, err := getIPInfo(ip)
	if err != nil {
		http.Error(w, "获取ip-api信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	address, err := getBaiduAddress(ip, baiduAK)
	if err != nil {
		http.Error(w, "获取百度定位失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderImage(w, ip, info, address)
}

func main() {
	baiduAK = os.Getenv("BAIDU_API_KEY")
	if baiduAK == "" {
		log.Fatal("请通过环境变量 BAIDU_API_KEY 提供百度API Key")
	}

	http.HandleFunc("/AQuVk4853gdCarY6ysbscNZCL4A7ndgK", handler)
	http.HandleFunc("/9uTWQKMJPcUuihF4LUtCtj48GkkRaZ82", handlerWithQuery)

	fmt.Println("Server running at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
