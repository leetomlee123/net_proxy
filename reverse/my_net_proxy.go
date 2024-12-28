package reverse

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// 监听的地址和端口
const proxyAddr = "0.0.0.0:8080"

// 上游代理地址（指向 goproxy）
const upstreamProxy = "127.0.0.1:4568"

func Init() {
	log.Println("Starting proxy server on", proxyAddr)
	listener, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 读取客户端请求的第一行数据
	reader := bufio.NewReader(clientConn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading request: %v", err)
		return
	}

	// 分析请求行，区分 HTTP 或 HTTPS
	if strings.HasPrefix(requestLine, "CONNECT") {
		handleHTTPS(clientConn, requestLine, reader)
	} else {
		handleHTTP(clientConn, requestLine, reader)
	}
}

func handleHTTPS(clientConn net.Conn, requestLine string, reader *bufio.Reader) {
	log.Printf("Handling HTTPS request: %s", requestLine)

	// 提取目标地址
	tokens := strings.Split(requestLine, " ")
	if len(tokens) < 2 {
		log.Println("Invalid CONNECT request")
		return
	}
	// targetAddr := tokens[1]

	// 与上游代理建立隧道连接
	upstreamConn, err := net.Dial("tcp", upstreamProxy)
	if err != nil {
		log.Printf("Error connecting to upstream proxy: %v", err)
		return
	}
	defer upstreamConn.Close()

	// 将 CONNECT 请求转发给上游代理
	_, err = fmt.Fprintf(upstreamConn, requestLine)
	if err != nil {
		log.Printf("Error sending CONNECT to upstream proxy: %v", err)
		return
	}

	// 读取上游代理的响应
	resp, err := http.ReadResponse(bufio.NewReader(upstreamConn), nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Upstream proxy CONNECT failed: %v", err)
		return
	}

	// 向客户端返回成功的 CONNECT 响应
	_, err = fmt.Fprintf(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		log.Printf("Error sending CONNECT response to client: %v", err)
		return
	}

	// 开始转发数据
	transferData(clientConn, upstreamConn)
}

func handleHTTP(clientConn net.Conn, requestLine string, reader *bufio.Reader) {
	log.Printf("Handling HTTP request: %s", requestLine)

	// 与上游代理建立连接
	upstreamConn, err := net.Dial("tcp", upstreamProxy)
	if err != nil {
		log.Printf("Error connecting to upstream proxy: %v", err)
		return
	}
	defer upstreamConn.Close()

	// 转发初始请求行和剩余的请求数据
	_, err = upstreamConn.Write([]byte(requestLine))
	if err != nil {
		log.Printf("Error forwarding request line: %v", err)
		return
	}

	_, err = io.Copy(upstreamConn, reader)
	if err != nil {
		log.Printf("Error forwarding request body: %v", err)
		return
	}

	// 转发上游代理的响应
	transferData(clientConn, upstreamConn)
}

func transferData(conn1, conn2 net.Conn) {
	done := make(chan struct{})

	// 双向转发
	go func() {
		io.Copy(conn1, conn2)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(conn2, conn1)
		done <- struct{}{}
	}()

	<-done
}
