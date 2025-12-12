// proxy.go
package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	// 连接目标（HTTPS）
	dst, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer dst.Close()

	// 响应 200 Connection Established
	w.WriteHeader(http.StatusOK)

	// Hijack 连接，双向透传
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		return
	}
	defer clientConn.Close()

	// 双向拷贝（goroutine 安全）
	go io.Copy(dst, clientConn)
	io.Copy(clientConn, dst)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// 记录监控信息（关键！）
	log.Printf("[HTTP] %s -> %s %s", r.RemoteAddr, r.Method, r.URL.String())

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 回写响应
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(">>> Received request: Method=%s, URL=%s, Host=%s", r.Method, r.URL.String(), r.Host)
		if r.Method == http.MethodConnect {
			log.Printf("[HTTPS] %s -> %s", r.RemoteAddr, r.Host)
			handleTunnel(w, r)
		} else {
			handleHTTP(w, r)
		}
	})

	log.Println("🚀 HTTP/HTTPS Proxy listening on :8989")
	log.Fatal(http.ListenAndServe(":8989", handler)) // ← 直接传 handler，不走 DefaultServeMux
}
