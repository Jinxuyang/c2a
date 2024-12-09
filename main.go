package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)
import "github.com/gin-gonic/gin"

var (
	config *Config
)

type Config struct {
	ThirdPartyURL string `json:"third_party_url"`
	Cookie        string `json:"cookie"`
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
	router.Use(Cors())
	router.POST("/v1/chat/completions", handler)

	port := "8080"
	log.Printf("Server listening on port %s", port)
	log.Fatal(router.Run(fmt.Sprintf(":%s", port)))
}

func handler(c *gin.Context) {
	// 解析 OpenAI 请求
	var openAIRequest OpenAIRequest
	if err := c.ShouldBindJSON(&openAIRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 转换为第三方请求格式
	thirdPartyRequest := convertToThirdPartyRequest(openAIRequest)

	// 将第三方请求转为 JSON
	requestBody, err := json.Marshal(thirdPartyRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal third-party request"})
		return
	}

	// 返回 OpenAI 响应
	c.Writer.Header().Set("Transfer-Encoding", "chunked") // 添加流式响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)

	if err := sendToThirdPartyAPI(requestBody, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response from third-party API"})
		return
	}
}

func convertToThirdPartyRequest(openAIRequest OpenAIRequest) ThirdPartyRequest {
	return ThirdPartyRequest{
		ChatSettings: struct {
			Model                        string  `json:"model"`
			Prompt                       string  `json:"prompt"`
			Temperature                  float64 `json:"temperature"`
			ContextLength                int     `json:"contextLength"`
			IncludeProfileContext        bool    `json:"includeProfileContext"`
			IncludeWorkspaceInstructions bool    `json:"includeWorkspaceInstructions"`
			EmbeddingsProvider           string  `json:"embeddingsProvider"`
		}{
			Model:                        "gpt-4o",
			Prompt:                       openAIRequest.Messages[0].Content,
			Temperature:                  openAIRequest.Temperature,
			ContextLength:                4096,     // example value, adjust as needed
			IncludeProfileContext:        true,     // example value, adjust as needed
			IncludeWorkspaceInstructions: true,     // example value, adjust as needed
			EmbeddingsProvider:           "openai", // example value, adjust as needed
		},
		Messages:      openAIRequest.Messages,
		CustomModelId: "",
	}
}

func sendToThirdPartyAPI(requestBody []byte, w http.ResponseWriter) error {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", config.ThirdPartyURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	//req.Header.Set("Cookie", "sb-gpt-kong-auth-token-code-verifier=%226933708d27ab1ccc44c0d4f4b0c21183fc73a3bc492972ec2f4829f1d3ad10321d12e2774380716b42fa6ba70f6344e494eacbfaf4b13fa8%22; digest=64a56da862e1b7b5dc9f5a34eca9c4ef51f9242f; cross_digest=848yv5S96resWlTpB7TCsAJGRguYGtbWuu7F3VSs73Z3LrjnGfiHSULSZRXyF%2FGn; sb-gpt-kong-auth-token=%7B%22access_token%22%3A%22eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjMmNmMDA0Ny00NDYzLTRlYWEtOGRjNi04NmVlODg2YWZkMzgiLCJhdWQiOiJhdXRoZW50aWNhdGVkIiwiZXhwIjoxNzMzNzM4MDUzLCJpYXQiOjE3MzM3MzQ0NTMsImVtYWlsIjoiamlueHV5YW5nQGtpbmdzb2Z0LmNvbSIsInBob25lIjoiIiwiYXBwX21ldGFkYXRhIjp7InByb3ZpZGVyIjoiZW1haWwiLCJwcm92aWRlcnMiOlsiZW1haWwiXX0sInVzZXJfbWV0YWRhdGEiOnsiZW1haWwiOiJqaW54dXlhbmdAa2luZ3NvZnQuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJwaG9uZV92ZXJpZmllZCI6ZmFsc2UsInN1YiI6ImMyY2YwMDQ3LTQ0NjMtNGVhYS04ZGM2LTg2ZWU4ODZhZmQzOCJ9LCJyb2xlIjoiYXV0aGVudGljYXRlZCIsImFhbCI6ImFhbDEiLCJhbXIiOlt7Im1ldGhvZCI6InBhc3N3b3JkIiwidGltZXN0YW1wIjoxNzMzMTM0ODAxfV0sInNlc3Npb25faWQiOiI1YzFkZGYxNi05YTYxLTRjNmQtOThiNi1jNWYyZjAyMWFlY2UiLCJpc19hbm9ueW1vdXMiOmZhbHNlfQ.CWU8FviI1mPRRGTd9tNr_uIx4ykgkEv8tI4PngeOqLg%22%2C%22token_type%22%3A%22bearer%22%2C%22expires_in%22%3A3600%2C%22expires_at%22%3A1733738053%2C%22refresh_token%22%3A%22juoNfZJi5z4g4pMtw6uU0A%22%2C%22user%22%3A%7B%22id%22%3A%22c2cf0047-4463-4eaa-8dc6-86ee886afd38%22%2C%22aud%22%3A%22authenticated%22%2C%22role%22%3A%22authenticated%22%2C%22email%22%3A%22jinxuyang%40kingsoft.com%22%2C%22email_confirmed_at%22%3A%222024-12-02T10%3A20%3A01.975762Z%22%2C%22phone%22%3A%22%22%2C%22confirmed_at%22%3A%222024-12-02T10%3A20%3A01.975762Z%22%2C%22last_sign_in_at%22%3A%222024-12-02T10%3A20%3A01.982702Z%22%2C%22app_metadata%22%3A%7B%22provider%22%3A%22email%22%2C%22providers%22%3A%5B%22email%22%5D%7D%2C%22user_metadata%22%3A%7B%22email%22%3A%22jinxuyang%40kingsoft.com%22%2C%22email_verified%22%3Afalse%2C%22phone_verified%22%3Afalse%2C%22sub%22%3A%22c2cf0047-4463-4eaa-8dc6-86ee886afd38%22%7D%2C%22identities%22%3A%5B%7B%22identity_id%22%3A%225198460f-0fe3-462e-a62d-707e061bc54e%22%2C%22id%22%3A%22c2cf0047-4463-4eaa-8dc6-86ee886afd38%22%2C%22user_id%22%3A%22c2cf0047-4463-4eaa-8dc6-86ee886afd38%22%2C%22identity_data%22%3A%7B%22email%22%3A%22jinxuyang%40kingsoft.com%22%2C%22email_verified%22%3Afalse%2C%22phone_verified%22%3Afalse%2C%22sub%22%3A%22c2cf0047-4463-4eaa-8dc6-86ee886afd38%22%7D%2C%22provider%22%3A%22email%22%2C%22last_sign_in_at%22%3A%222024-12-02T10%3A20%3A01.972566Z%22%2C%22created_at%22%3A%222024-12-02T10%3A20%3A01.972601Z%22%2C%22updated_at%22%3A%222024-12-02T10%3A20%3A01.972601Z%22%2C%22email%22%3A%22jinxuyang%40kingsoft.com%22%7D%5D%2C%22created_at%22%3A%222024-12-02T10%3A20%3A01.958574Z%22%2C%22updated_at%22%3A%222024-12-09T08%3A54%3A13.112188Z%22%2C%22is_anonymous%22%3Afalse%7D%7D")
	req.Header.Set("Cookie", config.Cookie)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := w.Write(buf[:n]) // 将数据写入客户端响应
			if writeErr != nil {
				return writeErr
			}
			// 刷新缓冲区，立即发送数据到客户端
			if f, ok := w.(http.Flusher); ok {
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
	return nil
}

func convertToOpenAIResponse(thirdPartyResponse ThirdPartyResponse) OpenAIResponse {
	return OpenAIResponse{
		Choices: []struct {
			Message Message `json:"message"`
		}{
			{Message: Message{Role: "assistant", Content: thirdPartyResponse.Content}},
		},
	}
}
