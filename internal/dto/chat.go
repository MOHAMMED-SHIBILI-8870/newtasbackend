// internal/dto/chat.go
package dto

type SendMessageRequest struct {
	ReceiverID uint   `json:"receiver_id"`
	BookingID  uint   `json:"booking_id"`
	Message    string `json:"message"`
}