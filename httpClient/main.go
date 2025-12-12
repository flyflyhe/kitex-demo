// proxy-test-client.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	// 👇 替换为你的代理地址
	proxyURL, err := url.Parse("http://127.0.0.1:8989")
	if err != nil {
		log.Fatal("Invalid proxy URL:", err)
	}

	// 创建带代理的 HTTP Client
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyURL(proxyURL),
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 20 * time.Second,
	}

	fmt.Println("🧪 Testing HTTP request via proxy...")
	resp1, err := client.Get("http://httpbin.org/get")
	if err != nil {
		log.Printf("❌ HTTP failed: %v", err)
	} else {
		defer resp1.Body.Close()
		body, _ := io.ReadAll(resp1.Body)
		fmt.Printf("✅ HTTP Status: %d\n", resp1.StatusCode)
		if len(body) > 200 {
			fmt.Println(string(body[:200]) + "...")
		} else {
			fmt.Println(string(body))
		}
	}

	fmt.Println("\n🧪 Testing HTTPS request via proxy...")
	resp2, err := client.Get("https://httpbin.org/get")
	if err != nil {
		log.Printf("❌ HTTPS failed: %v", err)
	} else {
		defer resp2.Body.Close()
		body, _ := io.ReadAll(resp2.Body)
		fmt.Printf("✅ HTTPS Status: %d\n", resp2.StatusCode)
		if len(body) > 200 {
			fmt.Println(string(body[:200]) + "...")
		} else {
			fmt.Println(string(body))
		}
	}
}
