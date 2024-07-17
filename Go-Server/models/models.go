package models

import (
	"github.com/gorilla/websocket"
)

type Client struct { // Client = Player
	Conn           *websocket.Conn // Connection 오브젝트
	ClientNickname string          // 플레이어 이름
	Score          int             // 현재 점수
	WriteChan      chan []byte     // 쓰기 채널 추가
}
type ClientJson struct {
	ClientNickname string `json:"clientNickname"` // 플레이어 닉네임
	Score          int    `json:"score"`          // 현재 점수
}

func ToClientJson(client *Client) *ClientJson {
	return &ClientJson{
		ClientNickname: client.ClientNickname,
	}
}

type Lobby struct {
	Clients   map[string]*Client  // 로비에 있는 client 목록
	Rooms     map[string]*Room    // 현재 방 목록
	Broadcast chan MessageWrapper // 방송: 방 목록 변경
}
type LobbyJson struct {
	Clients map[string]*ClientJson `json:"clients"` // 로비에 있는 client 목록
	Rooms   map[string]*RoomJson   `json:"rooms"`   // 현재 방 목록
}

func ToLobbyJson(lobby *Lobby) *LobbyJson {
	ret := &LobbyJson{
		Clients: make(map[string]*ClientJson),
		Rooms:   make(map[string]*RoomJson),
	}

	for clientUUID, client := range lobby.Clients {
		ret.Clients[clientUUID] = ToClientJson(client)
	}

	for roomUUID, room := range lobby.Rooms {
		ret.Rooms[roomUUID] = ToRoomJson(room)
	}

	return ret
}

type Room struct {
	RoomName        string              // 방 제목
	MaxParticipants int                 // 최대 참가자 수
	Clients         map[string]*Client  // 참가자
	Broadcast       chan MessageWrapper // 방송될 메시지
	MasterClient    *Client             // 방장
	RoomState       RoomState           // 게임 상태
}
type RoomJson struct {
	RoomName        string                 `json:"roomName"`        // 방 제목
	MaxParticipants int                    `json:"maxParticipants"` // 최대 참가자 수
	Clients         map[string]*ClientJson `json:"clients"`         // 참가자
	MasterClient    *ClientJson            `json:"masterClient"`    // 방장
	RoomState       RoomState              `json:"roomState"`       // 게임 상태
}

func ToRoomJson(room *Room) *RoomJson {
	ret := &RoomJson{
		RoomName:        room.RoomName,
		MaxParticipants: room.MaxParticipants,
		Clients:         make(map[string]*ClientJson),
		MasterClient:    ToClientJson(room.MasterClient),
		RoomState:       room.RoomState,
	}

	for clientUUID, client := range room.Clients {
		ret.Clients[clientUUID] = ToClientJson(client)
	}

	return ret
}

type RoomState struct {
	IsInGame            bool    `json:"isInGame"`            // 게임 중 여부
	IsCountingDown      bool    `json:"isCountingDown"`      // 카운트다운 관련 여부
	CountdownTime       int     `json:"countdownTime"`       // 카운트다운 시간
	IsOnPreparedImage   bool    `json:"isOnPreparedImage"`   // 이미지 사전 준비 중 여부
	PreparedSpriteSheet []byte  `json:"preparedSpriteSheet"` // 사전 준비된 퀴즈 스프라이트 시트
	Quizzes             []*Quiz `json:"quizzes"`             // 한 게임의 전체 퀴즈 목록
	CompletedQuizzesNum int     `json:"completedQuizzesNum"` // 완전한 정답이 발생한 퀴즈 개수
	CurrentQuizIndex    int     `json:"currentQuizIndex"`    // 현재 퀴즈 이미지
	CurrentQuiz         *Quiz   `json:"currentQuiz"`         // 현재 퀴즈
}

type Quiz struct {
	Keywords        []Keyword `json:"keywords"`        // 퀴즈 정답 키워드
	GuessedKeywords []Keyword `json:"guessedKeywords"` // 현재까지 맞춘 정답 키워드(tag_name 값)
	RemainingTime   int       `json:"remainingTime"`   // 현재 퀴즈 남은 시간
	TimerChannel    chan bool `json:"-"`               // 퀴즈 타이머 채널
}

type Keyword struct {
	CategoryId   int      `json:"category_id"`   // 카테고리 id
	CategoryName string   `json:"category_name"` // 카테고리 이름
	TagId        int      `json:"tag_id"`        // 태그 id
	TagName      string   `json:"tag_name"`      // 태그 이름(공식 정답)
	Answers      []string `json:"tag_answers"`   // 정답으로 사용될 문자열
}

type CategoryTagPair struct {
	CategoryId int `json:"category_id"`
	TagId      int `json:"tag_id"`
}
