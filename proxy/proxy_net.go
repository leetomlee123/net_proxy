package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"tidy/websocket"

	"gopkg.in/elazarl/goproxy.v1"
)

func GetUserAccount(token string) string {
	// Your existing implementation of GetUserAccount
	return ""
}
func readFile(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return data, nil
}

func Init() {
	// dir, err := os.Getwd()
	// if err != nil {
	// 	log.Fatalf("Failed to get working directory: %v", err)
	// }
	// fmt.Println("Current working directory:", dir)

	// // 拼接证书文件路径
	// certFile := filepath.Join("../", "certs", "ca.pem")
	// certKeyFile := filepath.Join("../", "certs", "ca.key.pem")

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
	caCert := `-----BEGIN CERTIFICATE-----
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
9jbBHhy6dl41xOpbhEaZIDBB+ZeaR2DkVA==
-----END CERTIFICATE-----
`
	caKey := `-----BEGIN PRIVATE KEY-----
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
TtDfK6eS/H7E8H8SLyW0yDC/1gyCMg==
-----END PRIVATE KEY-----
`
	setCA([]byte(caCert), []byte(caKey))
	proxy := goproxy.NewProxyHttpServer()

	// Enable HTTPS interception with MITM
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// Intercept HTTP and HTTPS requests and print request details
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			fmt.Printf("Request Method: %s, URL: %s\n", r.Method, r.URL.String())
			if r.URL.Host == "m.jkvugn.bar" {
				r.URL.Scheme = "https"
				r.URL.Host = "m.zzyi4cf7z8.cn:443"

				fmt.Printf("Rewritten URL: %s\n", r.URL.String())
			}
			// var udtauth12Value string

			// // Check if URL starts with specific prefix
			// if strings.HasPrefix(r.URL.String(), "https://m.zzyi4cf7z8.cn:443/tuijian") {
			// 	for name, values := range r.Header {
			// 		for _, value := range values {
			// 			fmt.Printf("Request Header: %s: %s\n", name, value)
			// 			if name == "Udtauth12" {
			// 				// Store the header value for later use (after response is received)
			// 				udtauth12Value = value
			// 			}
			// 		}
			// 	}
			// }
			// // Attach the stored value to context for use in the response phase
			// ctx.UserData = udtauth12Value

			// WebSocket notifications based on headers
			for name, values := range r.Header {
				for _, value := range values {
					fmt.Printf("Request Header: %s: %s\n", name, value)
					if strings.Contains(value, "PHPSESSID=") {
						value = strings.ReplaceAll(value, "PHPSESSID=", "")
						err := websocket.MyWebSocket.WriteMessage(1, []byte("tianxia://"+value))
						if err != nil {
							log.Println("Error writing to websocket:", err)
						}
					}
				}
			}

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
			if resp != nil && resp.Body != nil {
				// Read the response body
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					// Reassign the body so it can be read again downstream
					resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

					// Print or process the response body content
					fmt.Printf("Response Body: %s\n", string(bodyBytes))

					// Retrieve the stored value from request phase
					if strings.HasPrefix(resp.Request.URL.String(), "https://m.zzyi4cf7z8.cn:443/tuijian") {

						// for name, values := range resp.Request.Header {
						// 	for _, value := range values {
						// 		fmt.Printf("Request Header: %s: %s\n", name, value)
						// 		if value == "Udtauth12" {
						// 			token = value
						// 		}
						// 	}
						// }

						var jsonResponse map[string]interface{}
						err = json.Unmarshal(bodyBytes, &jsonResponse)
						if err != nil {
							log.Println("Error parsing JSON response:", err)
							return resp
						}
						if code, ok := jsonResponse["code"].(float64); ok && code != 0 {
							modifiedResponse := map[string]interface{}{
								"code": 0,
								"data": map[string]interface{}{
									"user": map[string]interface{}{
										"username":          "lix",
										"upuid":             "0",
										"uid":               "3799193",
										"regtime":           "2024-09-19 09:54:32",
										"score":             "10000.0000",
										"rebate_count_show": true,
										"rebate_count":      "0",
										"new_read_count":    "0",
									},
									"readCfg": map[string]interface{}{
										"check_score": 0,
										"user_score":  1.1,
									},
									"infoView": map[string]interface{}{
										"num":    "1",
										"score":  0,
										"rest":   0,
										"status": 1,
									},
									"tips": "通知：收徒奖励高，平均1元/徒弟，月收徒奖励万元",
								},
							}

							// Convert the modified response to JSON
							modifiedResponseBytes, err := json.Marshal(modifiedResponse)
							if err != nil {
								log.Println("Error marshaling modified response:", err)
								return resp
							}

							// Replace the original response body with the modified content
							resp.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedResponseBytes))
							resp.ContentLength = int64(len(modifiedResponseBytes))
							resp.Header.Set("Content-Type", "application/json")

							return resp
						}

						// Extract the required fields
						userData, ok := jsonResponse["data"].(map[string]interface{})["user"].(map[string]interface{})
						if !ok {
							log.Println("Error extracting user data from response")
							return resp
						}

						username := userData["username"].(string)
						uid := userData["uid"].(string)
						score := userData["score"].(string)
						token := ""
						if resp != nil && resp.Request != nil {
							// Access the request headers
							headers := resp.Request.Header

							// Check if the Udtauth12 header exists
							if values, ok := headers["Udtauth12"]; ok {
								// Iterate through the header values (in case there are multiple values)
								for _, value := range values {
									fmt.Printf("Udtauth12 header value: %s\n", value)
									token = value // Store the Udtauth12 value in token variable or use as needed
								}
							} else {
								// Udtauth12 header not found
								fmt.Println("Udtauth12 header not found in the request")
							}
						}
						// Assemble and send WebSocket message after getting the response
						message := fmt.Sprintf("kele://username=%s&uid=%s&score=%s&token=%s", username, uid, score, token)

						fmt.Println(message)

						go func() {
							defer func() {
								if r := recover(); r != nil {
									log.Printf("Recovered from panic in WebSocket message goroutine: %v\n", r)
								}
							}()

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
			return resp
		})
	// Enable verbose logging for debugging
	proxy.Verbose = true

	// Start the proxy server on port 4568
	log.Println("Proxy server started on :4568")
	log.Fatal(http.ListenAndServe(":4568", proxy))
}
