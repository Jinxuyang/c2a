package model

type ChatbotUIRequest struct {
	ChatSettings struct {
		Model                        string  `json:"model"`
		Prompt                       string  `json:"prompt"`
		Temperature                  float64 `json:"temperature"`
		ContextLength                int     `json:"contextLength"`
		IncludeProfileContext        bool    `json:"includeProfileContext"`
		IncludeWorkspaceInstructions bool    `json:"includeWorkspaceInstructions"`
		EmbeddingsProvider           string  `json:"embeddingsProvider"`
	} `json:"chatSettings"`
	Messages      []Message `json:"messages"`
	CustomModelId string    `json:"customModelId"`
}

func ConvertToChatbotUIRequest(request OpenAIRequest) ChatbotUIRequest {
	return ChatbotUIRequest{
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
			Prompt:                       request.Messages[0].Content,
			Temperature:                  request.Temperature,
			ContextLength:                4096,     // example value, adjust as needed
			IncludeProfileContext:        true,     // example value, adjust as needed
			IncludeWorkspaceInstructions: true,     // example value, adjust as needed
			EmbeddingsProvider:           "openai", // example value, adjust as needed
		},
		Messages:      request.Messages,
		CustomModelId: "",
	}
}

type ChatbotUIResponse struct {
	Content string `json:"content"`
}
