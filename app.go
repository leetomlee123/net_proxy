package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"tidy/proxy"
	"tidy/reverse"
	"tidy/websocket"
)

func handleConnect(w http.ResponseWriter, r *http.Request) {
	// 建立与目标服务器的连接
	target, err := url.Parse("http://" + r.Host)
	if err != nil {
		http.Error(w, "Invalid target", http.StatusBadRequest)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

func main() {

	http.HandleFunc("CONNECT", handleConnect)
	log.Fatal(http.ListenAndServe(":8080", nil))
	go func() {
		proxy.Init()

	}()
	go func() {
		websocket.Init()
	}()
	reverse.Init()

}
