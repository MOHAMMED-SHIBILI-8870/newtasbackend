package dto

type UpdateGuideProfileInput struct {
	Bio         string `json:"bio"`
	Experience  int    `json:"experience"`
	Languages   string `json:"languages"`
	IsAvailable bool   `json:"is_available"`
}