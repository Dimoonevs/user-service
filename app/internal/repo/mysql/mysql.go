package mysql

import (
	"database/sql"
	"flag"
	"github.com/Dimoonevs/user-service/app/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"log"
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
