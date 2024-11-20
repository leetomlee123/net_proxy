package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"tidy/websocket"


	"flag"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/elazarl/goproxy.v1"
)


func Init() {

	port := flag.String("p", "4568", "Port for the proxy server") // -p flag with default value "4568"
	flag.Parse()                                                  // Parse the flags
	setCA([]byte(caCert), []byte(caKey))

	proxy := goproxy.NewProxyHttpServer()
	// com.example.wx_reader
	// Enable HTTPS interception with MITM

	// myHandler := &MyReqHandler{}
	// proxy.OnRequest().Do(myHandler)

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// targetPaths := []string{
	// 	"yunonline/v1/gold",
	// 	"ttz/yn/queryUserCode",
	// 	"ttz/uaction/getArticleListkkk",
	// 	"v1/user/login",
	// 	"tuijian",
	// 	"yeipad/user",
	// }
	// var patterns []string
	// for _, path := range targetPaths {
	// 	patterns = append(patterns, regexp.QuoteMeta(path))
	// }

	// // 生成只要包含目标路径即可匹配的正则
	// combinedPattern := "(" + strings.Join(patterns, "|") + ")"
	// re := regexp.MustCompile(combinedPattern)

	// proxy.OnRequest(goproxy.UrlMatches(re)).HandleConnect(goproxy.AlwaysMitm)

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic:", r)
				}
			}()

			if resp == nil || resp.Body == nil {
				fmt.Println("Response or response body is nil")
				return nil // or return a default response or error as needed
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return resp
			}

			// Reset the response body
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			// Launch handlers in goroutines
			go func() {
				handleResponse(resp)

			}()

			go func() {
				handleXiaoyueyueResponse(resp, body)

			}()

			go func() {
				handleKeleResponse(resp, body)

			}()

			go func() {
				handleYuerResponse(resp, body)

			}()

			go func() {
				handleBaishitongResponse(resp, body)

			}()

			go func() {
				handleYoumiResponse(resp, body)

			}()

			// fmt.Println("Before returning response " + string(body))
			return resp
		})

	// Enable verbose logging for debugging
	proxy.Verbose = true

	// Start the proxy server on port 4568
	fmt.Printf("Starting GoProxy on :%s\n", *port)
	if err := http.ListenAndServe(":"+*port, proxy); err != nil {
		fmt.Println("Error starting proxy server:", err)
		os.Exit(1)
	}
}

func handleXiaoyueyueResponse(resp *http.Response, bodyBytes []byte) {

	// http://1730875136te.7-6-7.cn/?inviteid=0
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle xiaoyueyue panic: %v\n", r)
		}
	}()
	// Filter out OPTIONS requests
	if resp.Request.Method == http.MethodOptions {
		log.Println("OPTIONS request - no further processing.")
		return
	}

	headers := resp.Request.Header
	// Check if the URL contains "tuijian"
	urlStr := resp.Request.URL.String()
	if strings.Contains(urlStr, "/yunonline/v1/gold") {
		var token string
		if values, ok := headers["Cookie"]; ok {
			for _, value := range values {
				if strings.Contains(value, "ysmuid") {
					// 找到 ysmuid 的值
					parts := strings.Split(value, "ysmuid=")
					if len(parts) > 1 {
						token = strings.Split(parts[1], ";")[0] // 取 ysmuid 值
						fmt.Printf("xiaoyueyue token header value: %s\n", token)
					}
					break
				}
			}
		} else {
			fmt.Println("Cookie header not found in the request")
		}

		// 提取 Referer 和 User-Agent
		var referer, ua string
		if refererValues, ok := headers["Referer"]; ok && len(refererValues) > 0 {
			referer = refererValues[0]
		} else {
			fmt.Println("Referer header not found")
		}

		if uaValues, ok := headers["User-Agent"]; ok && len(uaValues) > 0 {
			ua = uaValues[0]
		} else {
			fmt.Println("User-Agent header not found")
		}
		username := getId(referer, token, ua)
		fmt.Println("Generated username:", username)
		headersMap := make(map[string]string)
		for key, values := range resp.Request.Header {
			if key == "Content-Length" || key == "Content-Type" {
				continue
			}
			// Assuming one value per key, but you could adapt this for multiple values
			headersMap[key] = values[0]
		}
		host := fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host)
		log.Printf("Request Host: %s\n", host)
		headersMap["hostX"] = host
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			log.Println("Error parsing URL:", err)
			return
		}

		// Retrieve specific GET parameters
		unionid := parsedURL.Query().Get("unionid")
		headersMap["unionid"] = unionid

		headersJSON, err := json.Marshal(headersMap)
		if err != nil {
			log.Println("Error marshaling headers to JSON:", err)
			return
		}
		// Assemble the WebSocket message
		message := fmt.Sprintf("xiaoyueyue://username=%s&params=%s&type=%s&token=%s&headers=%s", username, "", "小阅阅", token, headersJSON)

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

func getId(url, token, ua string) string {
	// 目标 URL

	// 发起请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	// 设置请求头，包括 User-Agent 和 Cookie
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Cookie", "ysmuid="+token)

	// 发起请求并获取响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Error loading document:", err)
		return ""
	}
	id := ""

	// 提取包含ID的文本
	doc.Find(".msg_con p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		id = strings.ReplaceAll(text, "我的id:", "")
	})
	return id
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

type Response struct {
	Code int `json:"code"`
	Data struct {
		User struct {
			Username string `json:"username"`

			UID string `json:"uid"`
		} `json:"user"`

		Tips string `json:"tips"`
	} `json:"data"`
}
type UserInfo struct {
	Nickname   string `json:"nickname"`
	Mobile     string `json:"mobile"`
	Avatar     string `json:"avatar"`
	FID        int64  `json:"fid"`
	Token      string `json:"token"`
	UserID     int64  `json:"user_id"`
	CreateTime int64  `json:"createtime"`
	ExpireTime int64  `json:"expiretime"`
	ExpiresIn  int64  `json:"expires_in"`
}

// Data 代表 JSON 响应中的数据部分
type Data struct {
	UserInfo UserInfo `json:"userinfo"`
}

// Response 代表完整的 JSON 响应结构
type BaishitongResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time string `json:"time"`
	Data Data   `json:"data"`
}
type YoumiResponse struct {
	Code    int       `json:"code"`
	Data    YouMiData `json:"data"`
	Message string    `json:"message"`
	Success bool      `json:"success"`
}

type YouMiData struct {
	FFB      string `json:"ffb"`
	Code     string `json:"code"`
	SJ       string `json:"sj"`
	StartNum string `json:"startNum"`
	EndNum   string `json:"endNum"`
	URL      string `json:"url"`
}

func handleYoumiResponse(resp *http.Response, bodyBytes []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle youmi login panic: %v\n", r)
		}
	}()
	// Filter out OPTIONS requests
	if resp.Request.Method == http.MethodOptions {
		log.Println("OPTIONS request - no further processing.")
		return
	}
	// http://hpage.ttianzhuan.cn/ttz/yn/userSignin?userShowId=499388
	// Check if the URL contains "tuijian"
	urlStr := resp.Request.URL.String()

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Error parsing URL:", err)
		return
	}

	// Retrieve specific GET parameters

	if strings.Contains(urlStr, "ttz/yn/queryUserCode") {
		userShowId := parsedURL.Query().Get("userShowId")
		http.NewRequest("GET", "http://hpage.ttianzhuan.cn/ttz/yn/userSignin?userShowId="+userShowId, nil)
	} else if strings.Contains(urlStr, "/ttz/uaction/getArticleListkkk") {
		var jsonResponse YoumiResponse
		err := json.Unmarshal(bodyBytes, &jsonResponse)
		if err != nil {
			log.Println("Error parsing JSON response:", err)
			return
		}

		// Extract the required fields from the response struct
		code := jsonResponse.Code
		if code == 200 && jsonResponse.Data.Code == "200" {

			username := parsedURL.Query().Get("str")
			token := parsedURL.Query().Get("token")

			headersMap := make(map[string]string)
			for key, values := range resp.Request.Header {
				if key == "Content-Length" || key == "Content-Type" {
					continue
				}
				// Assuming one value per key, but you could adapt this for multiple values
				headersMap[key] = values[0]
			}

			headersJSON, err := json.Marshal(headersMap)
			if err != nil {
				log.Println("Error marshaling headers to JSON:", err)
				return
			}
			paramMap := make(map[string]string)
			paramMap["startNumber"] = jsonResponse.Data.StartNum
			paramMap["keys"] = parsedURL.Query().Get("keys")

			paramJSON, err := json.Marshal(paramMap)
			if err != nil {
				log.Println("Error marshaling params to JSON:", err)
				return
			}

			// Assemble the WebSocket message
			message := fmt.Sprintf("youmi://username=%s&params=%s&type=%s&token=%s&headers=%s", username, paramJSON, "有米", token, headersJSON)

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
}
func handleBaishitongResponse(resp *http.Response, bodyBytes []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle baishitong login panic: %v\n", r)
		}
	}()
	// Filter out OPTIONS requests
	if resp.Request.Method == http.MethodOptions {
		log.Println("OPTIONS request - no further processing.")
		return
	}

	// Check if the URL contains "tuijian"
	if strings.Contains(resp.Request.URL.String(), "/v1/user/login") {
		var jsonResponse BaishitongResponse
		err := json.Unmarshal(bodyBytes, &jsonResponse)
		if err != nil {
			log.Println("Error parsing JSON response:", err)
			return
		}

		// Extract the required fields from the response struct
		code := jsonResponse.Code
		if code == 1 {
			username := jsonResponse.Data.UserInfo.Nickname
			token := jsonResponse.Data.UserInfo.Token
			headersMap := make(map[string]string)
			for key, values := range resp.Request.Header {
				if key == "Content-Length" || key == "Content-Type" {
					continue
				}
				// Assuming one value per key, but you could adapt this for multiple values
				headersMap[key] = values[0]
			}
			headersJSON, err := json.Marshal(headersMap)
			if err != nil {
				log.Println("Error marshaling headers to JSON:", err)
				return
			}
			// Assemble the WebSocket message
			message := fmt.Sprintf("baishitong://username=%s&uid=%s&type=%s&token=%s&headers=%s", username, "", "百事通", token, headersJSON)

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
}
func handleKeleResponse(resp *http.Response, bodyBytes []byte) {
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
		// score := userData.Score

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
		if username == "" {
			return
		}
		// Assemble the WebSocket message
		message := fmt.Sprintf("kele://username=%s&uid=%s&type=%s&token=%s", username, uid, "可乐", token)

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
func handleYuerResponse(resp *http.Response, bodyBytes []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle fangzhao kele  panic: %v\n", r)
		}
	}()
	// Filter out OPTIONS requests
	if resp.Request.Method == http.MethodOptions {
		log.Println("OPTIONS request - no further processing.")
		return
	}

	// Check if the URL contains "tuijian"
	if strings.Contains(resp.Request.URL.String(), "yeipad/user") {
		token := ""

		headers := resp.Request.Header

		// Check if the Udtauth12 header exists
		if values, ok := headers["Cookie"]; ok {
			for _, value := range values {
				if strings.Contains(value, "PHPSESSID") {
					token = strings.Split(value, "=")[1]
					fmt.Printf("kele PHPSESSID header value: %s\n", token)
					break
				}
			}
		} else {
			fmt.Println("PHPSESSID header not found in the request")
		}
		html := string(bodyBytes)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Fatal(err)
		}
		username := ""
		// Use the provided CSS selector to find the first <p> element
		selection := doc.Find("body > div.content > div > div.user-main > div.user-info > div > p:nth-child(1)")
		selection.Each(func(i int, s *goquery.Selection) {
			username = s.Text()
			fmt.Println("Found username:", username)
		})
		headersMap := make(map[string]string)
		for key, values := range resp.Request.Header {
			if key == "Content-Length" || key == "Content-Type" || key == "X-Forwarded-For" || key == "X-Real-Ip" {
				continue
			}
			// Assuming one value per key, but you could adapt this for multiple values
			headersMap[key] = values[0]
		}

		headersJSON, err := json.Marshal(headersMap)
		if err != nil {
			log.Println("Error marshaling headers to JSON:", err)
			return
		}
		// Assemble the WebSocket message
		message := fmt.Sprintf("kele://username=%s&uid=%s&type=%s&token=%s&headers=%s", username, "", "鱼儿", token, headersJSON)

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
func handleResponse(resp *http.Response) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handle tianxia  panic: %v\n", r)
		}
	}()
	urlStr := resp.Request.URL.String()
	// Check if the request URL matches the specific page
	if strings.Contains(urlStr, "read.index.html") {
	OuterLoop:
		for name, values := range resp.Request.Header {
			for _, value := range values {
				fmt.Printf("Request Header: %s: %s\n", name, value)
				// Look for the PHPSESSID token
				if strings.Contains(value, "PHPSESSID=") {
					split := strings.Split(value, ";")
					token := ""
					for _, vv := range split {
						// Trim any leading or trailing whitespace in each cookie
						vv = strings.TrimSpace(vv)

						// Check if this part contains "PHPSESSID="
						if strings.HasPrefix(vv, "PHPSESSID=") {
							// Replace only "PHPSESSID=" in this specific cookie
							token = strings.ReplaceAll(vv, "PHPSESSID=", "")
							break // Stop looping once we find the PHPSESSID token
						}
					}
					fmt.Printf("PHPSESSID Token: %s\n", token)

					account := ""

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
					baseURL := "https://api.nicevoice.nyc.mn/admin/wxCode"
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

// dir, err := os.Getwd()
// if err != nil {
// 	log.Fatalf("Failed to get working directory: %v", err)
// }
// fmt.Println("Current working directory:", dir)

// // 拼接证书文件路径
// certFile := filepath.Join("", "certs", "ca.crt")
// certKeyFile := filepath.Join("", "certs", "ca.key.pem")

// // 读取证书文件
// caCert, err := readFile(certFile)
// if err != nil {
// 	log.Fatalf(err.Error())
// }

// // 读取证书密钥文件
// caKey, err := readFile(certKeyFile)
// if err != nil {
// 	log.Fatalf(err.Error())
// }
