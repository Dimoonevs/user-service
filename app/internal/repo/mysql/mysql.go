package mysql

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/Dimoonevs/user-service/app/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"sync"
)

type Storage struct {
	db *sql.DB
}

var (
	mysqlConnectionString = flag.String("SQLConnPassword", "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4,utf8", "DB connection")
	storage               *Storage
	once                  sync.Once
)

func initMySQLConnection() {
	dbConn, err := sql.Open("mysql", *mysqlConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	dbConn.SetMaxIdleConns(0)

	storage = &Storage{
		db: dbConn,
	}
}

func GetConnection() *Storage {
	once.Do(func() {
		initMySQLConnection()
	})

	return storage
}

func (s *Storage) SaveUserData(userData models.UserData) (int, error) {
	query := `INSERT INTO users (email, password_hash, verification_token, is_verified) VALUES (?, ?, ?, 0)`

	result, err := s.db.Exec(query, userData.Email, userData.Password, userData.Code)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *Storage) GetUserByEmail(email string) (*models.UserData, error) {
	query := `SELECT id, is_verified, verification_token, password_hash FROM users WHERE email = ?`

	userData := &models.UserData{
		Email: email,
	}
	row := s.db.QueryRow(query, email)

	if err := row.Scan(&userData.ID, &userData.IsVerify, &userData.Code, &userData.Password); err != nil {
		logrus.Errorf("Cannot get code by email: %v", err)
		return nil, err
	}
	return userData, nil
}

func (s *Storage) UpdateVerifyCode(userID int, code string) error {
	query := `UPDATE users SET verification_token = ? WHERE id = ?`

	_, err := s.db.Exec(query, code, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) VerifyUser(email string) error {
	query := `UPDATE users SET is_verified = 1 WHERE email = ?`

	_, err := s.db.Exec(query, email)
	if err != nil {
		logrus.Errorf("Cannot update verefy user: %v", err)
		return err
	}
	return nil
}

func (s *Storage) IsVerifying(email string) bool {
	query := `SELECT is_verified FROM users WHERE email = ?`

	var isVerified bool
	row := s.db.QueryRow(query, email)
	if err := row.Scan(&isVerified); err != nil {
		logrus.Errorf("Cannot get verified user: %v", err)
		return false
	}
	return isVerified
}
func (s *Storage) ChangeDataUser(email, password string, userID int) error {
	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if email != "" {
		updates = append(updates, "email = ?")
		args = append(args, email)
	}
	if password != "" {
		updates = append(updates, "password_hash = ?")
		args = append(args, password)
	}

	if len(updates) == 0 {
		return fmt.Errorf("Nothin entered...")
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, userID)

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Storage) SetUserSettings(userID int, settings models.UserSettings) error {
	query := `INSERT INTO user_ai_settings (user_id, token, gpt_model, whisper_model, tts_model, name) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, userID, settings.AIToken, settings.GPTModel, settings.WhisperModel, settings.TTSModel, settings.Name)
	if err != nil {
		logrus.Errorf("Cannot set user settings: %v", err)
		return err
	}
	return nil
}

func (s *Storage) GetUserSettings(userID int) ([]*models.UserSettings, error) {
	query := `SELECT id, user_id, token, gpt_model, whisper_model, tts_model, name FROM user_ai_settings WHERE user_id = ?`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		logrus.Errorf("Cannot get user settings: %v", err)
		return nil, err
	}
	defer rows.Close()

	var settingsList []*models.UserSettings

	for rows.Next() {
		var settings models.UserSettings
		if err := rows.Scan(&settings.ID, &settings.UserID, &settings.AIToken, &settings.GPTModel, &settings.WhisperModel, &settings.TTSModel, &settings.Name); err != nil {
			logrus.Errorf("Cannot scan user settings: %v", err)
			return nil, err
		}
		settingsList = append(settingsList, &settings)
	}

	if err := rows.Err(); err != nil {
		logrus.Errorf("Error iterating user settings: %v", err)
		return nil, err
	}

	return settingsList, nil
}

func (s *Storage) UpdateUserSettings(userID int, settings models.UserSettings) error {
	query := "UPDATE user_ai_settings SET "
	args := []interface{}{}
	updates := []string{}

	if settings.AIToken != "" {
		updates = append(updates, "token = ?")
		args = append(args, settings.AIToken)
	}
	if settings.GPTModel != "" {
		updates = append(updates, "gpt_model = ?")
		args = append(args, settings.GPTModel)
	}
	if settings.WhisperModel != "" {
		updates = append(updates, "whisper_model = ?")
		args = append(args, settings.WhisperModel)
	}
	if settings.TTSModel != "" {
		updates = append(updates, "tts_model = ?")
		args = append(args, settings.TTSModel)
	}
	if settings.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, settings.Name)
	}

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ") + " WHERE user_id = ? AND id = ?"
	args = append(args, userID, settings.ID)

	_, err := s.db.Exec(query, args...)
	return err
}
