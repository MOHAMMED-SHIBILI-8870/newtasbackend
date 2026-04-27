package usecase

import (
	"backend/internal/config"
	"backend/internal/entity"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

//generate 6 DIGIT OTP

func GenerateOTP() (string,error) {
	n, err := rand.Int(rand.Reader,big.NewInt(900000))
	if err != nil{
		return "",err
	}
	otp := 100000 + n.Int64()
	return fmt.Sprintf("%06d",otp),nil
}

//OTP hashing

func HashOTP(otp string)string{
	hash := sha256.Sum256([]byte(otp))
	return  hex.EncodeToString(hash[:])
}

//create OTP with hashing save DB

func CreateOTP(db *gorm.DB,userID uint,purpose string,expiryMinutes int) (string,error){
	otp ,err := GenerateOTP()
	if err != nil{
		return  "",err
	}

	otpHash := HashOTP(otp)

	tx :=db.Begin()

	if err :=tx.Model(&entity.OTP{}).Where("user_id = ? AND purpose = ? AND is_used = false ",userID,purpose).
	Update("is_used",true).Error;err !=nil{
		tx.Rollback()
		return  "",err
	}

	updatedVersion  := entity.OTP{
		UserID: userID,
		OTPCode: otpHash,
		Purpose: purpose,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(expiryMinutes)),
		IsUsed: false,
	}

	if err := tx.Create(&updatedVersion).Error; err != nil {
		tx.Rollback()
		return  "",err
	}

	tx.Commit()
	return otp,nil
}


//verify OTP
func VerifyOTP(userId uint, otp, purpose string) (bool, error) {
	err := config.DB.Transaction(func(dt *gorm.DB) error {
		var entry entity.OTP
		hashedOTP := HashOTP(otp)

		if err := dt.Where(
			"user_id = ? AND purpose = ? AND otp_code = ? AND is_used = false AND expires_at > ?",
			userId, purpose, hashedOTP, time.Now(),
		).
			Order("created_at DESC").
			First(&entry).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired OTP")
			}
			return err
		}

		if err := dt.Model(&entry).Update("is_used", true).Error; err != nil {
			return err
		}

		if purpose == "signup" {
			if err := dt.Model(&entity.User{}).
				Where("id = ?", userId).
				Update("is_verified", true).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}