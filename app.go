package main

import (
	"tidy/proxy"
	"tidy/websocket"
)

func main() {
	go func() {
		websocket.Init()
	}()
	proxy.Init()
}
