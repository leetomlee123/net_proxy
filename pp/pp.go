package pp

import (
	"tidy/proxy"
	"tidy/websocket"
)



func Start() {
	go func() {
		websocket.Init()
	}()
	proxy.Init()
}