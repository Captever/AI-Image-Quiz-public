package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"websocket/constants"
	"websocket/db"
	"websocket/game"
	"websocket/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// 작업 소요 시간 측정 함수
func MeasureTime(start time.Time, requester, processName string) {
	elapsed := time.Since(start)
	log.Printf("%s %s 소요 시간: %.6fs", requester, processName, elapsed.Seconds())
}

func (s *Server) SendDataToClient(client *models.Client, message json.RawMessage) {
	client.WriteChan <- message
}

func (s *Server) ConnectToMaster(ws *websocket.Conn, clientNickname string) {
	clientUUID := uuid.New().String()
	client := &models.Client{
		Conn:           ws,
		ClientNickname: clientNickname,
		WriteChan:      make(chan []byte, 256),
	}

	go s.writePump(client)

	// log.Printf("Current Client(%s) Info: %v", clientUUID, client)

	s.JoinLobby(clientUUID, client) // MasterServer에 연결됐다면 자동으로 로비 입장

	// 클라이언트에게 현 client의 정보를 보내서 기억할 수 있도록
	jsonData := json.RawMessage(`{"clientUUID":"` + clientUUID + `"}`)
	mw := models.NewMessageWrapper("OnConnectedToMaster", clientUUID, jsonData)
	s.SendDataToClient(client, MarshalMwData(mw))

	s.UpdateLobby()
}

func (s *Server) JoinLobby(clientUUID string, client *models.Client) {
	s.Lobby.Clients[clientUUID] = client
	log.Printf("Client(%s) joined lobby", clientUUID)
}

func (s *Server) CreateRoom(roomName string, maxParticipants int, clientUUID string) {
	// TODO: 방 Id가 잘못되진 않았는지, 이미 존재하는 방은 아닌지 validation
	roomUUID := uuid.New().String()
	s.Lobby.Rooms[roomUUID] = &models.Room{
		RoomName:        roomName,
		MaxParticipants: maxParticipants,
		Clients:         make(map[string]*models.Client),
		Broadcast:       make(chan models.MessageWrapper),
		MasterClient:    s.Lobby.Clients[clientUUID], // 방을 만든 사람이 최초의 방장
	}
	log.Printf("Client(%s) created room(%s)", clientUUID, roomName)

	go s.handleRoomBroadcast(roomUUID)

	s.JoinRoom(roomUUID, clientUUID) // 방을 만들었다면 자동으로 방에 입장

	// 클라이언트에게 현 room의 정보를 보내서 기억할 수 있도록
	jsonData := json.RawMessage(`{"roomUUID":"` + roomUUID + `"}`)
	mw := models.NewMessageWrapper("OnCreatedRoom", clientUUID, jsonData)
	s.SendDataToClient(s.Lobby.Clients[clientUUID], MarshalMwData(mw))
}
func (s *Server) ResetRoomState(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	room.RoomState = models.RoomState{}
}

func (s *Server) JoinRoom(roomUUID string, clientUUID string) {
	s.Lobby.Rooms[roomUUID].Clients[clientUUID] = s.Lobby.Clients[clientUUID] // 방에 클라이언트 추가

	log.Printf("Client(%s) joined room(%s)", clientUUID, roomUUID)

	jsonData := json.RawMessage(`{"roomUUID":"` + roomUUID + `"}`)
	mw := models.NewMessageWrapper("OnJoinedRoom", clientUUID, jsonData)
	s.SendDataToClient(s.Lobby.Clients[clientUUID], MarshalMwData(mw))

	s.UpdateRoom(roomUUID)
	s.UpdateLobby()

	s.ResetStartCountdown(roomUUID)
}

// 게임 시작 카운트다운 리셋
func (s *Server) ResetStartCountdown(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]
	if room.RoomState.IsCountingDown {
		// 기존 카운트다운 취소
		room.RoomState.IsCountingDown = false
	}
	s.CheckGameStartCondition(roomUUID)
}

func (s *Server) LeftRoom(roomUUID string, clientUUID string) {
	currRoom := s.Lobby.Rooms[roomUUID]

	delete(s.Lobby.Rooms[roomUUID].Clients, clientUUID) // 방에서 클라이언트 제거

	// 방에 아무도 남아있지 않다면 DropRoom 실행
	if len(currRoom.Clients) == 0 {
		s.DropRoom(roomUUID)
	} else {
		// 방장이 나갔으면 새로운 방장을 지정
		if currRoom.MasterClient == s.Lobby.Clients[clientUUID] {
			for _, newMasterClient := range currRoom.Clients {
				currRoom.MasterClient = newMasterClient
				break
			}
			log.Printf("New master for room(%s) is client(%s)", currRoom.RoomName, currRoom.MasterClient.ClientNickname)
		}
		s.UpdateRoom(roomUUID)
		s.UpdateLobby()

		s.CheckGameStartCondition(roomUUID)
	}

	log.Printf("Client(%s) left room(%s)", clientUUID, roomUUID)

}

func (s *Server) DropRoom(roomUUID string) {
	delete(s.Lobby.Rooms, roomUUID) // 방 목록에서 방 제거
	log.Printf("room(%s) is dropped", roomUUID)

	s.UpdateLobby()
}

func (s *Server) SendChatMessage(clientUUID, roomUUID, content string) {
	room := s.Lobby.Rooms[roomUUID]

	// 현재 게임 중이라면 퀴즈 메시지로 이동
	if room.RoomState.IsInGame {
		start := time.Now()
		s.SendQuizMessage(clientUUID, roomUUID, content)
		MeasureTime(start, clientUUID, "퀴즈 메시지 처리")
		return
	}

	senderNickname := s.Lobby.Clients[clientUUID].ClientNickname
	message := senderNickname + ": " + content

	jsonData := json.RawMessage(`{"message":"` + message + `"}`)
	s.BroadcastToRoom(roomUUID, "OnRecievedChatMessage", jsonData)

	log.Printf("get room(%s) a message: %s", roomUUID, message)
}
func (s *Server) SendQuizMessage(clientUUID, roomUUID, content string) {
	senderNickname := s.Lobby.Clients[clientUUID].ClientNickname
	message := senderNickname + ": " + content

	jsonData := json.RawMessage(`{"message":"` + message + `"}`)
	s.BroadcastToRoom(roomUUID, "OnRecievedChatMessage", jsonData)

	s.CheckAnswer(roomUUID, clientUUID, content)
}
func (s *Server) SendSystemMessage(roomUUID, content string) {
	message := "[시스템] " + content

	jsonData := json.RawMessage(`{"message":"` + message + `"}`)
	s.BroadcastToRoom(roomUUID, "OnRecievedSystemMessage", jsonData)

	log.Printf("get room(%s) a system message: %s", roomUUID, message)
}

func (s *Server) CheckGameStartCondition(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	// 최대 인원의 과반수 이상이면 게임 시작 카운트다운
	if len(room.Clients) > room.MaxParticipants/2 {
		if !room.RoomState.IsCountingDown {
			s.StartCountdown(roomUUID, constants.START_COUNTDOWN_TIME)
		}
	} else {
		if room.RoomState.IsCountingDown {
			room.RoomState.IsCountingDown = false
			s.BroadcastToRoom(roomUUID, "OnCancelledStartCountdown", nil)
		}
	}
}

func (s *Server) StartCountdown(roomUUID string, countdown int) {
	room := s.Lobby.Rooms[roomUUID]
	room.RoomState.IsCountingDown = true
	room.RoomState.CountdownTime = countdown

	// 초기 카운트다운 값을 즉시 전송
	s.SendStartCountdown(roomUUID, countdown)

	// 이미지 처리 시작
	if !room.RoomState.IsOnPreparedImage && room.RoomState.PreparedSpriteSheet == nil {
		log.Println("Start process: Preparing Game Images")
		go s.PrepareGameSpriteSheet(roomUUID)
	}

	go func() {
		// 0초가 됐을 때까지 신호를 줘야 카운트다운이 종료됨
		for {
			time.Sleep(1 * time.Second)
			room.RoomState.CountdownTime--

			s.SendStartCountdown(roomUUID, room.RoomState.CountdownTime)

			if room.RoomState.CountdownTime <= 0 {
				break
			}
		}

		if room.RoomState.IsCountingDown {
			room.RoomState.IsCountingDown = false
			s.EnableGameStart(roomUUID)
		}
	}()
}
func (s *Server) SendStartCountdown(roomUUID string, countdown int) {
	jsonData := json.RawMessage(`{"startCountdownTime":"` + strconv.Itoa(countdown) + `"}`)
	s.BroadcastToRoom(roomUUID, "OnUpdatedStartCountdown", jsonData)
}

func (s *Server) EnableGameStart(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	if room == nil {
		log.Printf("Unknown RoomUUID(%v): Unable to enable game start", roomUUID)
		return
	}

	mw := models.NewMessageWrapper("OnEnabledStart", "", nil)
	s.SendDataToClient(room.MasterClient, MarshalMwData(mw))
}

func (s *Server) StartGame(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]
	room.RoomState.IsInGame = true

	log.Printf("Game Started on '%v'\n", room.RoomName)

	// 준비된 sprite sheet가 있는지 확인
	if room.RoomState.PreparedSpriteSheet == nil {
		log.Printf("Error: Sprite sheet not prepared for room %s", room.RoomName)
		return
	}

	// Base64로 인코딩
	spriteSheetBase64 := base64.StdEncoding.EncodeToString(room.RoomState.PreparedSpriteSheet)

	// 클라이언트에 전송
	jsonData := json.RawMessage(`{"spriteSheet":"` + spriteSheetBase64 + `"}`)
	s.BroadcastToRoom(roomUUID, "OnStartedGame", jsonData)

	room.RoomState.CurrentQuizIndex = 0
	s.StartQuiz(roomUUID)
}

func (s *Server) PrepareGameSpriteSheet(roomUUID string) {
	start := time.Now()
	defer MeasureTime(start, roomUUID, "이미지 불러오기")

	room := s.Lobby.Rooms[roomUUID]

	// 중복 실행이 되지 않도록 bool 변수 지정
	room.RoomState.IsOnPreparedImage = true
	defer func() {
		room.RoomState.IsOnPreparedImage = false
	}()

	// 9개의 랜덤 이미지 URL 가져오기
	imageURLs, imageKeywords, err := db.GetRandomImages()
	if err != nil {
		log.Printf("Error getting random images: %v\n", err)
		return
	}

	// Sprite sheet 생성
	spriteSheetBytes, err := game.CreateSpriteSheet(imageURLs[:])
	if err != nil {
		log.Printf("Error creating sprite sheet: %v", err)
		return
	}

	// 준비된 sprite sheet 저장
	room.RoomState.PreparedSpriteSheet = spriteSheetBytes

	// 퀴즈 생성
	fmt.Println("==QuizList==")
	room.RoomState.Quizzes = make([]*models.Quiz, constants.IMAGE_COUNT)
	for i, keywords := range imageKeywords {
		quiz := &models.Quiz{
			Keywords:        make([]models.Keyword, len(keywords)),
			GuessedKeywords: make([]models.Keyword, 0),
			RemainingTime:   constants.QUIZ_TIMER_TIME,
			TimerChannel:    make(chan bool),
		}
		copy(quiz.Keywords, keywords[:])
		room.RoomState.Quizzes[i] = quiz

		for _, keyword := range quiz.Keywords {
			fmt.Printf("%s: %s, ", keyword.CategoryName, keyword.TagName)
		}
		fmt.Printf("\n")
	}
}

func (s *Server) StartQuiz(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	room.RoomState.CurrentQuiz = room.RoomState.Quizzes[room.RoomState.CurrentQuizIndex]

	// 현재 퀴즈의 카테고리 이름 추출
	categoryNames := make([]string, constants.CATEGORY_COUNT)
	for i, keyword := range room.RoomState.CurrentQuiz.Keywords {
		categoryNames[i] = keyword.CategoryName
	}

	// 카테고리 이름을 클라이언트에 전달
	jsonData := MarshalMwData(map[string][]string{
		"categoryNames": categoryNames,
	})
	s.BroadcastToRoom(roomUUID, "OnStartedQuiz", jsonData)

	// 퀴즈 카운트다운 시작
	s.StartQuizCountdown(roomUUID, constants.QUIZ_COUNTDOWN_TIME)
}

func (s *Server) StartQuizCountdown(roomUUID string, countdown int) {
	room := s.Lobby.Rooms[roomUUID]
	room.RoomState.IsCountingDown = true
	room.RoomState.CountdownTime = countdown

	// 초기 카운트다운 값을 즉시 전송
	s.SendQuizCountdown(roomUUID, countdown)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		room.RoomState.CountdownTime--

		if room.RoomState.CountdownTime <= 0 {
			room.RoomState.IsCountingDown = false
			s.SendQuizCountdown(roomUUID, 0)
			s.ShowQuizImage(roomUUID)
			return
		}

		s.SendQuizCountdown(roomUUID, room.RoomState.CountdownTime)
	}
}
func (s *Server) SendQuizCountdown(roomUUID string, countdown int) {
	jsonData := json.RawMessage(`{"quizCountdownTime":"` + strconv.Itoa(countdown) + `"}`)
	s.BroadcastToRoom(roomUUID, "OnUpdatedQuizCountdown", jsonData)
}

func (s *Server) ShowQuizImage(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]
	jsonData := json.RawMessage(`{"currentQuizIndex":"` + strconv.Itoa(room.RoomState.CurrentQuizIndex) + `"}`)
	s.BroadcastToRoom(roomUUID, "OnShownQuizImage", jsonData)

	go s.RunQuizTimer(roomUUID)
}

func (s *Server) RunQuizTimer(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	// 초기 타이머 값을 즉시 전송
	s.SendQuizRemainingTime(roomUUID, room.RoomState.CurrentQuiz.RemainingTime)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-room.RoomState.CurrentQuiz.TimerChannel:
			// 타이머 중지 신호를 받았을 때
			return
		case <-ticker.C:
			room.RoomState.CurrentQuiz.RemainingTime--
			if room.RoomState.CurrentQuiz.RemainingTime <= 0 {
				s.EndQuiz(roomUUID)
				return
			}
			s.SendQuizRemainingTime(roomUUID, room.RoomState.CurrentQuiz.RemainingTime)
		}
	}
}
func (s *Server) SendQuizRemainingTime(roomUUID string, remainingTime int) {
	jsonData := json.RawMessage(`{"remainingTime":"` + strconv.Itoa(remainingTime) + `"}`)
	s.BroadcastToRoom(roomUUID, "OnUpdatedQuizRemainingTime", jsonData)
}

func (s *Server) EndQuiz(roomUUID string) {
	room := s.Lobby.Rooms[roomUUID]

	// 타이머 중지
	close(room.RoomState.CurrentQuiz.TimerChannel)

	// 현재 퀴즈의 정답 목록 추출
	keywords := make([]string, constants.CATEGORY_COUNT)
	for i, keyword := range room.RoomState.CurrentQuiz.Keywords {
		keywords[i] = keyword.TagName
	}

	// 정답 목록을 클라이언트에 전달
	jsonData := MarshalMwData(map[string][]string{
		"keywords": keywords,
	})
	s.BroadcastToRoom(roomUUID, "OnRevealedAnswer", jsonData)

	// 다음 퀴즈로 이동
	room.RoomState.CurrentQuizIndex++

	// 9개의 모든 퀴즈를 사용했거나 5개의 완전 통과 이미지가 발생하면 게임 종료
	if room.RoomState.CurrentQuizIndex >= constants.MAX_QUIZ_COUNT || room.RoomState.CompletedQuizzesNum >= constants.CORRECTED_QUIZ_COUNT {
		s.EndGame(roomUUID)
	} else {
		s.StartQuiz(roomUUID)
	}
}

func (s *Server) EndGame(roomUUID string) {
	start := time.Now()
	defer MeasureTime(start, roomUUID, "합산 점수 산출")

	room := s.Lobby.Rooms[roomUUID]
	room.RoomState.IsInGame = false

	scores := make(map[string]int)
	for clientUUID, client := range room.Clients {
		score := client.Score
		scores[client.ClientNickname] = score

		// 현재 점수 값을 서버에 반영한 후 값 초기화
		db.UpdateUserScore(clientUUID, strconv.Itoa(score))
		client.Score = 0

		log.Printf("User(%v) score: %d", clientUUID, score)
	}

	jsonData := MarshalMwData(scores)
	s.BroadcastToRoom(roomUUID, "OnEndedGame", jsonData)

	// 현재의 RoomState는 초기화
	s.ResetRoomState(roomUUID)

	// 게임 시작이 가능한 조건인지 다시 확인
	go s.CheckGameStartCondition(roomUUID)
}

func (s *Server) CheckAnswer(roomUUID, clientUUID, answer string) {
	room := s.Lobby.Rooms[roomUUID]

	// 이미 맞춘 정답이라면 스킵
	for _, guessedKeyword := range room.RoomState.CurrentQuiz.GuessedKeywords {
		if contains(guessedKeyword.Answers, answer) {
			return
		}
	}

	for _, keyword := range room.RoomState.CurrentQuiz.Keywords {
		if contains(keyword.Answers, answer) {
			room.RoomState.CurrentQuiz.GuessedKeywords = append(room.RoomState.CurrentQuiz.GuessedKeywords, keyword)

			s.OnOccuredCorrectAnswer(roomUUID, clientUUID, keyword)

			// 3개의 정답 키워드를 모두 맞췄다면 완전한 정답 개수 추가 및 퀴즈 종료
			if len(room.RoomState.CurrentQuiz.GuessedKeywords) == constants.CATEGORY_COUNT {
				room.RoomState.CompletedQuizzesNum++
				s.EndQuiz(roomUUID)
			}
			return
		}
	}
}
func (s *Server) OnOccuredCorrectAnswer(roomUUID, clientUUID string, answer models.Keyword) {
	room := s.Lobby.Rooms[roomUUID]
	client := room.Clients[clientUUID]
	category_name := answer.CategoryName
	answer_keyword := answer.TagName

	// TODO: 문제 당 5점으로 가정(추후 시간 비례해서 점수가 줄어들도록 하기)
	client.Score += constants.MAX_SCORE_PER_ANSWER
	log.Printf("User(%v) get score +5", clientUUID)

	jsonData := MarshalMwData(map[string]string{"clientNickname": client.ClientNickname, "category_name": category_name, "keyword": answer_keyword})
	s.BroadcastToRoom(roomUUID, "OnOccuredCorrectAnswer", jsonData)

	// 정답 관련 시스템 메시지 출력
	message := client.ClientNickname + "님이 카테고리(" + category_name + ")의 정답을 맞히셨습니다: " + answer_keyword
	s.SendSystemMessage(roomUUID, message)
}

func (s *Server) BroadcastToRoom(roomUUID, action string, jsonData json.RawMessage) {
	room := s.Lobby.Rooms[roomUUID]

	if room == nil {
		log.Printf("Undefined RoomUUID(%v): Unable to broadcast\n", roomUUID)
		return
	}

	mw := models.MessageWrapper{
		Action: action,
		Data:   jsonData,
	}

	room.Broadcast <- mw
}

func (s *Server) UpdateLobby() {
	jsonData := MarshalMwData(models.ToLobbyJson(s.Lobby))
	mw := models.MessageWrapper{
		Action:     "OnUpdatedLobby",
		ClientUUID: "",
		Data:       jsonData,
	}

	s.Lobby.Broadcast <- mw // 다른 클라이언트에게 전파
}
func (s *Server) UpdateRoom(roomUUID string) {
	currRoom := s.Lobby.Rooms[roomUUID]

	jsonData := MarshalMwData(models.ToRoomJson(currRoom))
	s.BroadcastToRoom(roomUUID, "OnUpdatedRoom", jsonData)
}

// MessageWrapper Data를 Marshal할 때 사용
func MarshalMwData(data interface{}) json.RawMessage {
	ret, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling data: %v", err)
	}
	return ret
}

// 문자열 배열에 문자열이 포함되어 있는지 확인하는 함수
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
