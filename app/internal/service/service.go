package service

import (
	"fmt"
	"github.com/Dimoonevs/user-service/app/internal/lib"
	"github.com/Dimoonevs/user-service/app/internal/models"
	"github.com/Dimoonevs/user-service/app/internal/repo/mysql"
	"github.com/Dimoonevs/user-service/app/pkg/jwt"
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

	if err = sendVerificationEmail(req.Email, code, "Confirmation of registration", "Enter this code to confirm your registration"); err != nil {
		logrus.Errorf("Failed to send verification email: %v", err)
		return 0, err
	}

	return userID, nil
}

func SendVerificationEmailAgain(email string) error {
	code := lib.GenerateSecureVerificationCode()

	err := mysql.GetConnection().SetCodeByEmail(email, code)
	if err != nil {
		return err
	}

	if err := sendVerificationEmail(email, code, "Confirmation of registration", "Enter this code to confirm your registration"); err != nil {
		logrus.Errorf("Failed to send verification email: %v", err)
		return err
	}
	return nil
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
	token, err := jwt.GenerateJWT(req.Email, userData.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func RequestResetPassword(email string) error {
	userData, err := mysql.GetConnection().GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to reset password for user with email: %s", email)
	}
	code := lib.GenerateSecureVerificationCode()

	if err = mysql.GetConnection().UpdateVerifyCode(userData.ID, code); err != nil {
		return fmt.Errorf("failed to reset password for user with email: %s", email)
	}

	if err = sendVerificationEmail(email, code, "Confirmation of reset password", "Enter this code to confirm reset password"); err != nil {
		logrus.Errorf("Failed to send verification email: %v", err)
		return err
	}
	return nil
}

func ConfirmResetPassword(req models.UsersReq) error {
	userData, err := mysql.GetConnection().GetUserByEmail(req.Email)
	if err != nil {
		return err
	}
	if req.Code != userData.Code {
		return fmt.Errorf("failed to verify code for user with email: %s", req.Email)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err = mysql.GetConnection().ChangeDataUser("", string(hashedPassword), userData.ID); err != nil {
		return err
	}
	return nil
}

// with token

func UserSettings(userID int, req models.UserSettings) error {
	if err := mysql.GetConnection().SetUserSettings(userID, req); err != nil {
		return err
	}
	return nil
}

func GetUserSettings(userID int) (settings []*models.UserSettings, err error) {
	settings, err = mysql.GetConnection().GetUserSettings(userID)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func UpdateUserSettings(userID int, settings models.UserSettings) error {
	if err := mysql.GetConnection().UpdateUserSettings(userID, settings); err != nil {
		return err
	}
	return nil
}
