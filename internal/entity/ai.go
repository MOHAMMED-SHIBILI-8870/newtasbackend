package entity

import "google.golang.org/genai"

type ChatRequest struct {
	// We pass the conversation history array directly in Gemini format
	Messages []*genai.Content `json:"messages"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}