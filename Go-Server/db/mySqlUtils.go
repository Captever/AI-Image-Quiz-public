package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"websocket/constants"

	"golang.org/x/crypto/bcrypt"
)

var mysqlDBClient *sql.DB

func InitMySQL() {
	var err error
	mysqlDBClient, err = sql.Open("mysql", constants.MYSQL_ENDPOINT)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = mysqlDBClient.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
}

func UpdateLastLogin(userId string) error {
	query := "UPDATE Users SET last_login = ? WHERE user_id = ?"
	_, err := mysqlDBClient.Exec(query, time.Now(), userId)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func SignupUser(username, password, email string) (int, string) {
	log.Println("회원가입 시도")

	// 비밀번호 해시화
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return DatabaseError, "Error hashing password"
	}

	// 사용자 존재 여부 확인
	var existingUser string
	err = mysqlDBClient.QueryRow("SELECT username FROM Users WHERE username = ?", username).Scan(&existingUser)
	if err == nil {
		log.Printf("Username already exists: %s", username)
		return UserExists, "Username already exists"
	} else if err != sql.ErrNoRows {
		log.Printf("Error querying database: %v", err)
		return DatabaseError, "Database error"
	}

	// 사용자 추가
	query := "INSERT INTO Users (username, password_hash, email) VALUES (?, ?, ?)"
	_, err = mysqlDBClient.Exec(query, username, hashedPassword, email)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
		return DatabaseError, "Database error"
	}

	return Success, "User signed up successfully"
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ValidateUser(username, password string) (int, string, string) {
	log.Println("유저 검증 시도")

	var storedPasswordHash string
	var userId string
	query := "SELECT user_id, password_hash FROM Users WHERE username = ?"
	err := mysqlDBClient.QueryRow(query, username).Scan(&userId, &storedPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Username not found: %s", username)
			return InvalidCredentials, "Invalid username.", ""
		}
		log.Printf("Error querying database: %v", err)
		return DatabaseError, "Database error", ""
	}

	if !CheckPasswordHash(password, storedPasswordHash) {
		return InvalidCredentials, "Incorrect password.", ""
	}

	return Success, "Login successful.", userId
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
