package service

import (
	"fmt"
	"github.com/Dimoonevs/user-service/app/internal/lib"
	"github.com/Dimoonevs/user-service/app/internal/models"
	"github.com/Dimoonevs/user-service/app/internal/repo/mysql"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(req models.UsersReq) (int, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	code := lib.GenerateSecureVerificationCode()
	userData := models.UserData{
		Email:    req.Email,
		Password: string(hashedPassword),
		Code:     code,
	}

	userID, err := mysql.GetConnection().SaveUserData(userData)
	if err != nil {
		return 0, err
	}

	if err := sendVerificationEmail(req.Email, code); err != nil {
		logrus.Errorf("Failed to send verification email: %v", err)
		return 0, err
	}

	return userID, nil
}

func VerifyCode(req models.UsersReq) error {
	userData, err := mysql.GetConnection().GetUserByEmail(req.Email)
	if err != nil {
		return err
	}
	if req.Code != userData.Code {
		return fmt.Errorf("failed to verify code for user with email: %s", req.Email)
	}

	if err = mysql.GetConnection().VerifyUser(req.Email); err != nil {
		return err
	}

	return nil
}

func LoginUser(req models.UsersReq) (string, error) {
	userData, err := mysql.GetConnection().GetUserByEmail(req.Email)
	if err != nil {
		return "", err
	}
	if !userData.IsVerify {
		return "", fmt.Errorf("failed to verify email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(req.Password))
	if err != nil {
		return "", fmt.Errorf("failed to compare password")
	}
	token, err := generateJWT(req.Email, userData.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
