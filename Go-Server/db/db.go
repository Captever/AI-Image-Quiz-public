//go:build prod
// +build prod

package db

// 상태 코드 정의
const (
	Success            = iota // 0
	InvalidCredentials        // 1
	DatabaseError             // 2
	UserExists                // 3
)

func InitDB() {
	InitMySQL()
	InitDynamoDB()
	InitS3()
}
