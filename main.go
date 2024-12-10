package main

import (
	"bytes"
	"c2a/middleware"
	"c2a/model"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)
import "github.com/gin-gonic/gin"

var (
	config *Config
)

type Config struct {
	ChatbotUIUrl string `json:"chatbot_ui_url"`
	Cookie       string `json:"cookie"`
}

func loadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}
	return nil
}

func main() {
	configFile := flag.String("c", "config.json", "Path to configuration file")
	flag.Parse()
	if err := loadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 使用 Gin 框架
	router := gin.Default()
	router.Use(middleware.Cors())
	router.POST("/v1/chat/completions", handler)
	router.GET("/v1/models", func(c *gin.Context) {
		res := "{\"object\":\"list\",\"data\":[{\"id\":\"gpt-4o\",\"object\":\"model\",\"created\":1686935002,\"owned_by\":\"openai\"}]}"
		c.Writer.WriteString(res)
	})

	port := "8080"
	log.Printf("Server listening on port %s", port)
	log.Fatal(router.Run(fmt.Sprintf(":%s", port)))
}

func handler(c *gin.Context) {
	// 解析 OpenAI 请求
	var openAIRequest model.OpenAIRequest
	if err := c.ShouldBindJSON(&openAIRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	requestBody, err := json.Marshal(model.ConvertToChatbotUIRequest(openAIRequest))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal third-party request"})
		return
	}

	if openAIRequest.Stream {
		c.Writer.Header().Set("Transfer-Encoding", "chunked") // 添加流式响应头
	}

	if err := sendToChatbotUIAPI(c, requestBody, openAIRequest.Stream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response from third-party API"})
		return
	}
}

func sendToChatbotUIAPI(c *gin.Context, requestBody []byte, isStream bool) error {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", config.ChatbotUIUrl, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Cookie", config.Cookie)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	if isStream {
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := c.Writer.Write(buf[:n]) // 将数据写入客户端响应
				if writeErr != nil {
					return writeErr
				}
				// 刷新缓冲区，立即发送数据到客户端
				if f, ok := c.Writer.(http.Flusher); ok {
					f.Flush() // 刷新数据到客户端
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		var res = model.ChatbotUIResponse{
			Content: string(body),
		}
		c.JSON(http.StatusOK, model.ConvertToOpenAIResponse(res))
	}

	return nil
}
