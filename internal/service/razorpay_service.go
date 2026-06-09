package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/razorpay/razorpay-go"
)

type RazorpayService struct {
	client *razorpay.Client
	secret string
}

func NewRazorpayService() *RazorpayService {
	key := os.Getenv("RAZORPAY_KEY_ID")
	secret := os.Getenv("RAZORPAY_KEY_SECRET")

	return &RazorpayService{
		client: razorpay.NewClient(key, secret),
		secret: secret,
	}
}

func (s *RazorpayService) CreateOrder(
	amount float64,
	receipt string,
) (map[string]interface{}, error) {

	data := map[string]interface{}{
		"amount":   int(amount * 100),
		"currency": "INR",
		"receipt":  receipt,
	}

	return s.client.Order.Create(data, nil)
}

func (s *RazorpayService) VerifySignature(
	orderID string,
	paymentID string,
	signature string,
) bool {

	payload := orderID + "|" + paymentID

	h := hmac.New(
		sha256.New,
		[]byte(s.secret),
	)

	h.Write([]byte(payload))

	expected := hex.EncodeToString(
		h.Sum(nil),
	)

	return expected == signature
}