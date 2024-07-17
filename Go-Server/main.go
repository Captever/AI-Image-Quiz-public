package main

import (
	"fmt"
	"os"
	"websocket/constants"
	"websocket/db"
	"websocket/server"
)

func main() {
	// 데이터베이스 연결
	db.InitDB()

	// 서버 연결
	srv := server.NewServer(constants.WS_PORT)
	fmt.Println("Starting server on port " + constants.WS_PORT)
	if err := srv.Start(); err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1) // 에러 발생 시 종료
	}
}
