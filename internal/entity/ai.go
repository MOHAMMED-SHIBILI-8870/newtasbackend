package entity

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}
