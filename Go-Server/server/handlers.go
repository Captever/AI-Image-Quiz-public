package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"websocket/db"
	"websocket/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Error upgrading connection: %v", err)
		return
	}
	defer ws.Close()

	s.handleMessages(ws)
}

func (s *Server) handleLobbyBroadcast() {
	for {
		mw := <-s.Lobby.Broadcast
		message := MarshalMwData(mw)
		for _, client := range s.Lobby.Clients {
			s.SendDataToClient(client, message)
		}
	}
}
func (s *Server) handleRoomBroadcast(roomUUID string) {
	for {
		mw := <-s.Lobby.Rooms[roomUUID].Broadcast
		message := MarshalMwData(mw)
		for _, client := range s.Lobby.Rooms[roomUUID].Clients {
			s.SendDataToClient(client, message)
		}
	}
}

func (s *Server) handleMessages(ws *websocket.Conn) {
	for {
		var mw models.MessageWrapper

		// 공통 메시지 구조체로 읽기
		err := ws.ReadJSON(&mw)
		if err != nil {
			// 연결이 끊어지면 루프 종료
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Connection closed:", err)
				break
			}
			log.Println("Error reading JSON:", err)
			break
		}

		fmt.Printf("Received: %s\n", mw)

		s.handleMessageWrapper(mw, ws)
	}
}

func (s *Server) handleMessageWrapper(mw models.MessageWrapper, ws *websocket.Conn) {
	// 메시지 타입에 따라 처리
	switch mw.Action {
	case "ConnectToMaster":
		var connectionMessage models.ConnectionMessage
		UnmarshalMwData(mw.Data, &connectionMessage)

		// fmt.Printf("Received clientNickname: %v\n", connectionMessage.ClientNickname)

		s.ConnectToMaster(ws, connectionMessage.ClientNickname)
	case "CreateRoom":
		start := time.Now()
		var createRoomMessage models.RoomMessage
		UnmarshalMwData(mw.Data, &createRoomMessage)

		// fmt.Printf("Received createRoom %v(max:%d)\n", createRoomMessage.RoomName, createRoomMessage.MaxParticipants)

		s.CreateRoom(createRoomMessage.RoomName, createRoomMessage.MaxParticipants, mw.ClientUUID)
		MeasureTime(start, mw.ClientUUID, "방 생성")
	case "JoinRoom":
		start := time.Now()
		var joinRoomMessage models.RoomMessage
		UnmarshalMwData(mw.Data, &joinRoomMessage)

		// fmt.Printf("Received joinRoom %v\n", joinRoomMessage.RoomUUID)

		s.JoinRoom(joinRoomMessage.RoomUUID, mw.ClientUUID)
		MeasureTime(start, mw.ClientUUID, "방 접속")
	case "LeftRoom":
		var leftRoomMessage models.RoomMessage
		UnmarshalMwData(mw.Data, &leftRoomMessage)

		// fmt.Printf("Received leftRoom %v\n", leftRoomMessage.RoomUUID)

		s.LeftRoom(leftRoomMessage.RoomUUID, mw.ClientUUID)
	case "SendChatMessage":
		var chatMessage models.ChatMessage
		UnmarshalMwData(mw.Data, &chatMessage)

		// fmt.Printf("Received SendChatMessage: %v\n", chatMessage.Content)

		s.SendChatMessage(mw.ClientUUID, chatMessage.RoomUUID, chatMessage.Content)
	case "StartGame":
		var startGameMessage models.RoomMessage
		UnmarshalMwData(mw.Data, &startGameMessage)

		s.StartGame(startGameMessage.RoomUUID)
	default:
		// 알 수 없는 메시지 타입 처리 로직
		fmt.Println("Unknown messageWrapper action:", mw.Action)
	}
}

// TODO: signin과 signup을 핸들러를 통합하기
func (s *Server) handleSigninSession(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.SigninRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer MeasureTime(start, req.Username, "로그인")

	statusCode, msg, userId := db.ValidateUser(req.Username, req.Password)
	log.Printf("Signin: (%d)%s", statusCode, msg)
	switch statusCode {
	case db.Success:
		// 사용자 ID를 가져와 세션 생성
		sessionId := uuid.New().String()
		err := db.CreateSession(userId, sessionId)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		response := models.SigninResponse{
			SessionUUID: sessionId,
			ClientUUID:  userId,
		}
		log.Printf("Succeed to create session: %v", sessionId)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	case db.InvalidCredentials:
		http.Error(w, msg, http.StatusUnauthorized)
	case db.DatabaseError:
		http.Error(w, msg, http.StatusInternalServerError)
	default:
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}
}

func (s *Server) handleSignoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.SignoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = db.DeleteSession(req.SessionUUID)
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	err = db.UpdateLastLogin(req.ClientUUID)
	if err != nil {
		http.Error(w, "Failed to update last login", http.StatusInternalServerError)
		return
	}

	log.Println("Session was signed out correctly")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout successful"))
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.SignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	statusCode, msg := db.SignupUser(req.Username, req.Password, req.Email)
	log.Printf("Signup: (%d)%s", statusCode, msg)
	switch statusCode {
	case db.Success:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
	case db.UserExists:
		http.Error(w, msg, http.StatusConflict)
	case db.DatabaseError:
		http.Error(w, msg, http.StatusInternalServerError)
	default:
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// 헬스 체크가 성공적으로 되었음을 알리기 위해 상태 코드 200 반환
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// MessageWrapper Data를 Unmarshal할 때 사용
func UnmarshalMwData(data json.RawMessage, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
	}
}
