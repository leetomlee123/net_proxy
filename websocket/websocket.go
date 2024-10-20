package websocket

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var MyWebSocket *websocket.Conn

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow any origin
    },
}

func handler(w http.ResponseWriter, r *http.Request) {
    var err error
    MyWebSocket, err = upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer MyWebSocket.Close()

    for {
        messageType, message, err := MyWebSocket.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }

        log.Printf("Received: %s", message)

        // Check if the message is "ping" to reply with "pong"
        if string(message) == "ping" {
            response := []byte("pong")
            if err := MyWebSocket.WriteMessage(messageType, response); err != nil {
                log.Println("Write error:", err)
                return
            }
            log.Println("Replied with pong")
        } else {
            // Echo the message back to the client
            if err := MyWebSocket.WriteMessage(messageType, message); err != nil {
                log.Println("Write error:", err)
                return
            }
        }
    }
}

func Init() {
    http.HandleFunc("/ws", handler)
    fmt.Println("WebSocket server started at :4567")
    log.Fatal(http.ListenAndServe(":4567", nil))
}
