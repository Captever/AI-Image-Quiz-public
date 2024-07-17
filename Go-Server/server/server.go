package server

import (
	"fmt"
	"net/http"
	"websocket/models"

	"github.com/gorilla/websocket"
)

type Server struct {
	Port     string
	Upgrader websocket.Upgrader
	Lobby    *models.Lobby
}

func NewServer(port string) *Server {
	s := &Server{
		Port: port,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Lobby: &models.Lobby{
			Clients:   make(map[string]*models.Client),
			Rooms:     make(map[string]*models.Room),
			Broadcast: make(chan models.MessageWrapper),
		},
	}

	go s.handleLobbyBroadcast()

	return s
}

func (s *Server) Start() error {
	http.HandleFunc("/ws", s.handleConnections)         // ws Path로 커넥션
	http.HandleFunc("/health", s.handleHealthCheck)     // health Path로 Health check
	http.HandleFunc("/signup", s.handleSignup)          // signup Path로 회원가입 관련 접근
	http.HandleFunc("/signin", s.handleSigninSession)   // signin Path로 로그인 관련 접근
	http.HandleFunc("/signout", s.handleSignoutSession) // signout Path로 로그아웃 관련 접근
	err := http.ListenAndServe(":"+s.Port, nil)         // 서버에 지정된 포트로 호스팅
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// 클라이언트에 메시지를 전달할 때 사용할 채널 시스템
func (s *Server) writePump(client *models.Client) {
	for {
		message, ok := <-client.WriteChan
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		client.Conn.WriteMessage(websocket.TextMessage, message)
	}
}
