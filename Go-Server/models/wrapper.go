package models

import (
	"encoding/json"
)

// 공통 메시지
type MessageWrapper struct {
	Action     string          `json:"action"`     // 수행해야 할 Action
	ClientUUID string          `json:"clientUUID"` // 요청자 클라이언트 ID
	Data       json.RawMessage `json:"data"`       // 실 데이터
}

func NewMessageWrapper(action string, clientUUID string, data json.RawMessage) *MessageWrapper {
	ret := &MessageWrapper{
		Action:     action,
		ClientUUID: clientUUID,
		Data:       data,
	}

	return ret
}

// 연결 관련(ex: ConnectToMaster)
type ConnectionMessage struct {
	ClientNickname string `json:"clientNickname"`
}

type RoomMessage struct {
	RoomName        string `json:"roomName"`        // 방 제목
	RoomUUID        string `json:"roomUUID"`        // 방 입장, 퇴장할 때 사용
	MaxParticipants int    `json:"maxParticipants"` // 방의 최대 참가자 수
}

type ChatMessage struct {
	RoomUUID string `json:"roomUUID"` // 메시지가 전달될 방의 식별자
	Content  string `json:"content"`  // 메시지 내용
}

// # request
type SigninRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type SignupRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type SignoutRequest struct {
	SessionUUID string `json:"sessionUUID"`
	ClientUUID  string `json:"clientUUID"`
	RoomUUID    string `json:"roomUUID"`
}

// # response
type SigninResponse struct {
	SessionUUID string `json:"sessionUUID"`
	ClientUUID  string `json:"clientUUID"`
}
