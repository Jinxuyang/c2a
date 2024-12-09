package model

type ThirdPartyRequest struct {
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

type ThirdPartyResponse struct {
	Content string `json:"content"`
}
