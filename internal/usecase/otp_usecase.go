package usecase

import (
	"os"

	verify "github.com/twilio/twilio-go/rest/verify/v2"
	"github.com/twilio/twilio-go"
)

type OTPUsecase struct {
	Client *twilio.RestClient
}

func NewOTPUsecase(client *twilio.RestClient) *OTPUsecase {
	return &OTPUsecase{Client: client}
}

// Send OTP
func (u *OTPUsecase) SendOTP(phone string) (string, error) {
	params := &verify.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel("sms")

	resp, err := u.Client.VerifyV2.CreateVerification(
		os.Getenv("VERIFY_SERVICE_SID"),
		params,
	)

	if err != nil {
		return "", err
	}

	return *resp.Status, nil
}

// Verify OTP
func (u *OTPUsecase) VerifyOTP(phone, code string) (bool, error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(phone)
	params.SetCode(code)

	resp, err := u.Client.VerifyV2.CreateVerificationCheck(
		os.Getenv("VERIFY_SERVICE_SID"),
		params,
	)

	if err != nil {
		return false, err
	}

	return resp.Status != nil && *resp.Status == "approved", nil
}