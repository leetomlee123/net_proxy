package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"tidy/websocket"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/elazarl/goproxy.v1"
)

func Init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	fmt.Println("Current working directory:", dir)

	// 拼接证书文件路径
	certFile := filepath.Join("", "certs", "ca.crt")
	certKeyFile := filepath.Join("", "certs", "ca.key.pem")

	// 读取证书文件
	caCert, err := readFile(certFile)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// 读取证书密钥文件
	caKey, err := readFile(certKeyFile)
	if err != nil {
		log.Fatalf(err.Error())
	}

	setCA([]byte(caCert), []byte(caKey))
	proxy := goproxy.NewProxyHttpServer()

	// Enable HTTPS interception with MITM
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// Intercept HTTP and HTTPS requests and print request details
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			fmt.Printf("Request Method: %s, URL: %s\n", r.Method, r.URL.String())

			clientIP := r.RemoteAddr
			fmt.Println(clientIP)

			// Add or append to the X-Forwarded-For header to include the client's real IP
			// If X-Forwarded-For already exists, append the new client IP
			if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
				r.Header.Set("X-Forwarded-For", prior+", "+clientIP)
			} else {
				r.Header.Set("X-Forwarded-For", clientIP)
			}

			// Optionally, you can also set X-Real-IP
			r.Header.Set("X-Real-IP", clientIP)

			// if r.URL.Host == "m.jkvugn.bar" {
			// 	r.URL.Scheme = "https"
			// 	r.URL.Host = "m.zzyi4cf7z8.cn:443"

			// 	fmt.Printf("Rewritten URL: %s\n", r.URL.String())
			// }
			// var udtauth12Value string

			// // Check if URL starts with specific prefix
			// if strings.HasPrefix(r.URL.String(), "https://m.zzyi4cf7z8.cn:443/tuijian") {
			for name, values := range r.Header {
				for _, value := range values {
					fmt.Printf("Request Header: %s: %s\n", name, value)
				}
			}
			// }
			// // Attach the stored value to context for use in the response phase
			// ctx.UserData = udtauth12Value

			// WebSocket notifications based on headers

			if r.Body != nil {
				bodyBytes, err := ioutil.ReadAll(r.Body)
				if err == nil {
					r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
					fmt.Printf("Request Body: %s\n", string(bodyBytes))
				}
			}

			return r, nil
		})
	// Intercept responses and print response details
	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// 处理错误
				return resp
			}

			// 重新设置响应体，以便客户端可以读取
			go handleHuyuegongxiangResponse(resp)
			go handleResponse(resp, body)

			// // 在 goroutine 中处理响应
			// go func() {
			// 	// 调用 tianxia 函数，处理响应
			// 	tianxia(resp)
			// }()
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			fmt.Println("before return resp")
			return resp
		})
	// Enable verbose logging for debugging
	proxy.Verbose = true

	// Start the proxy server on port 4568
	log.Println("Proxy server started on :4568")
	log.Fatal(http.ListenAndServe(":4568", proxy))
}

func GetUserAccount(token string) string {
	// 目标 URL
	url := "http://huyuegongxiang.2024.k9981.top/read.index.html"

	// 发起请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	// 设置请求头，包括 User-Agent 和 Cookie
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; PGT-AN00 Build/HONORPGT-AN00; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/126.0.6478.188 Mobile Safari/537.36 XWEB/1260183 MMWEBSDK/20240501 MMWEBID/207 MicroMessenger/8.0.50.2701(0x2800325B) WeChat/arm64 Weixin NetType/WIFI Language/zh_CN ABI/arm64")
	req.Header.Set("Cookie", "PHPSESSID="+token)

	// 发起请求并获取响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}

	// 解析 HTML 内容，提取账号
	// 正则表达式查找账号
	re := regexp.MustCompile(`账号:\s*<strong class="am-text-danger">(0+)</strong>`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		fmt.Println("Account number:", matches[1])
		return matches[1]
	} else {
		fmt.Println("Account number not found.")
		return ""
	}
}
func readFile(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return data, nil
}

// Print or process the response body content
// fmt.Printf("Response Body: %s\n", string(bodyBytes))

// Retrieve the stored value from request phase
// if strings.HasPrefix(resp.Request.URL.String(), "https://m.zzyi4cf7z8.cn:443/tuijian") {

// 	// for name, values := range resp.Request.Header {
// 	// 	for _, value := range values {
// 	// 		fmt.Printf("Request Header: %s: %s\n", name, value)
// 	// 		if value == "Udtauth12" {
// 	// 			token = value
// 	// 		}
// 	// 	}
// 	// }

// 	var jsonResponse map[string]interface{}
// 	err = json.Unmarshal(bodyBytes, &jsonResponse)
// 	if err != nil {
// 		log.Println("Error parsing JSON response:", err)
// 		return resp
// 	}
// 	if code, ok := jsonResponse["code"].(float64); ok && code != 0 {
// 		modifiedResponse := map[string]interface{}{
// 			"code": 0,
// 			"data": map[string]interface{}{
// 				"user": map[string]interface{}{
// 					"username":          "无无有歌",
// 					"upuid":             "0",
// 					"uid":               "3799193",
// 					"regtime":           "2024-09-19 09:54:32",
// 					"score":             "10000.0000",
// 					"rebate_count_show": true,
// 					"rebate_count":      "0",
// 					"new_read_count":    "0",
// 				},
// 				"readCfg": map[string]interface{}{
// 					"check_score": 0,
// 					"user_score":  1.1,
// 				},
// 				"infoView": map[string]interface{}{
// 					"num":    "1",
// 					"score":  0,
// 					"rest":   0,
// 					"status": 1,
// 				},
// 				"tips": "通知：收徒奖励高，平均1元/徒弟，月收徒奖励万元",
// 			},
// 		}

// 		// Convert the modified response to JSON
// 		modifiedResponseBytes, err := json.Marshal(modifiedResponse)
// 		if err != nil {
// 			log.Println("Error marshaling modified response:", err)
// 			return resp
// 		}

// 		// Replace the original response body with the modified content
// 		resp.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedResponseBytes))
// 		resp.ContentLength = int64(len(modifiedResponseBytes))
// 		resp.Header.Set("Content-Type", "application/json")

// 		return resp
// 	}

// 	// Extract the required fields
// 	userData, ok := jsonResponse["data"].(map[string]interface{})["user"].(map[string]interface{})
// 	if !ok {
// 		log.Println("Error extracting user data from response")
// 		return resp
// 	}

// 	username := userData["username"].(string)
// 	uid := userData["uid"].(string)
// 	score := userData["score"].(string)
// 	token := ""
// 	if resp != nil && resp.Request != nil {
// 		// Access the request headers
// 		headers := resp.Request.Header

// 		// Check if the Udtauth12 header exists
// 		if values, ok := headers["Udtauth12"]; ok {
// 			// Iterate through the header values (in case there are multiple values)
// 			for _, value := range values {
// 				fmt.Printf("Udtauth12 header value: %s\n", value)
// 				token = value // Store the Udtauth12 value in token variable or use as needed
// 			}
// 		} else {
// 			// Udtauth12 header not found
// 			fmt.Println("Udtauth12 header not found in the request")
// 		}
// 	}
// 	// Assemble and send WebSocket message after getting the response
// 	message := fmt.Sprintf("kele://username=%s&uid=%s&score=%s&token=%s", username, uid, score, token)

// 	fmt.Println(message)

// 	go func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				log.Printf("Recovered from panic in WebSocket message goroutine: %v\n", r)
// 			}
// 		}()

// 		err := websocket.MyWebSocket.WriteMessage(1, []byte(message))
// 		if err != nil {
// 			log.Println("Error writing to websocket:", err)
// 		} else {
// 			fmt.Printf("WebSocket message sent: %s\n", message)
// 		}

//		}()
//	}
type Response struct {
	Code int `json:"code"`
	Data struct {
		User struct {
			Username        string `json:"username"`
			UpUID           string `json:"upuid"`
			UID             string `json:"uid"`
			RegTime         string `json:"regtime"`
			Score           string `json:"score"`
			RebateCountShow bool   `json:"rebate_count_show"`
			RebateCount     string `json:"rebate_count"`
			NewReadCount    string `json:"new_read_count"`
		} `json:"user"`
		ReadCfg struct {
			CheckScore int     `json:"check_score"`
			UserScore  float64 `json:"user_score"`
		} `json:"readCfg"`
		InfoView struct {
			Num    string  `json:"num"`
			Score  float64 `json:"score"`
			Rest   int     `json:"rest"`
			Status int     `json:"status"`
		} `json:"infoView"`
		Tips string `json:"tips"`
	} `json:"data"`
}

func handleResponse(resp *http.Response, bodyBytes []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle kele tuijian panic: %v\n", r)
		}
	}()
	// Filter out OPTIONS requests
	if resp.Request.Method == http.MethodOptions {
		log.Println("OPTIONS request - no further processing.")
		return
	}

	// Check if the URL contains "tuijian"
	if strings.Contains(resp.Request.URL.String(), "tuijian") {
		var jsonResponse Response
		err := json.Unmarshal(bodyBytes, &jsonResponse)
		if err != nil {
			log.Println("Error parsing JSON response:", err)
			return
		}

		// Extract the required fields from the response struct
		userData := jsonResponse.Data.User
		username := userData.Username
		uid := userData.UID
		score := userData.Score

		// Extract the token from the headers
		token := ""
		if resp != nil && resp.Request != nil {
			headers := resp.Request.Header

			// Check if the Udtauth12 header exists
			if values, ok := headers["Udtauth12"]; ok {
				for _, value := range values {
					fmt.Printf("Udtauth12 header value: %s\n", value)
					token = value // Store the Udtauth12 value in token variable
				}
			} else {
				fmt.Println("Udtauth12 header not found in the request")
			}
		}

		// Assemble the WebSocket message
		message := fmt.Sprintf("kele://username=%s&uid=%s&score=%s&token=%s", username, uid, score, token)

		// Send WebSocket message in a separate goroutine
		fmt.Println(message)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in WebSocket message goroutine: %v\n", r)
				}
			}()

			// Write the message to WebSocket
			err := websocket.MyWebSocket.WriteMessage(1, []byte(message))
			if err != nil {
				log.Println("Error writing to websocket:", err)
			} else {
				fmt.Printf("WebSocket message sent: %s\n", message)
			}
		}()
	}
}
func handleHuyuegongxiangResponse(resp *http.Response) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle tianxia  panic: %v\n", r)
		}
	}()
	// Check if the request URL matches the specific page
	if resp.Request.URL.String() == "http://huyuegongxiang.2024.k9981.top/read.index.html" {
	OuterLoop:
		for name, values := range resp.Request.Header {
			for _, value := range values {
				fmt.Printf("Request Header: %s: %s\n", name, value)
				// Look for the PHPSESSID token
				if strings.Contains(value, "PHPSESSID=") {
					token := strings.ReplaceAll(value, "PHPSESSID=", "")
					account := ""

					fmt.Println("get token:", token)

					// Parse the HTML response body to extract the account
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						log.Fatal("Error parsing HTML document:", err)
					}

					// Find the account information within the HTML document
					doc.Find("span").Each(func(i int, s *goquery.Selection) {
						// Check if the text contains the word "账号"
						if strings.Contains(s.Text(), "账号") {
							// Extract the account from a `strong.am-text-danger` element
							account = s.Find("strong.am-text-danger").Text()
							fmt.Printf("账号: %s\n", account)
						}
					})

					// Construct the target URL with the extracted token and account
					baseURL := "http://127.0.0.1:1234/admin/wxCode"
					fullURL := fmt.Sprintf("%s?code=%s&tag=%s", baseURL, token, account)

					// Set up basic authentication
					username := "alex"
					password := "1qaz2wsx"
					auth := username + ":" + password
					encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

					// Create a new GET request with Basic authentication
					req, err := http.NewRequest("GET", fullURL, nil)
					if err != nil {
						fmt.Println("Error creating request:", err)
						break OuterLoop
					}
					req.Header.Add("Authorization", "Basic "+encodedAuth)

					// Send the request
					client := &http.Client{}
					resp1, err := client.Do(req)
					if err != nil {
						fmt.Println("Error sending request:", err)
						break OuterLoop
					}
					defer resp1.Body.Close()

					// Handle the response (for now just print out a message)
					fmt.Println("Request successfully sent with token and account")
					break OuterLoop
				}
			}
		}
	}
}
