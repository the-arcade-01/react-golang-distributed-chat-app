package models

type MessageType string

const (
	JoinMsgType  MessageType = "JOIN"
	LeaveMsgType MessageType = "LEAVE"
	ChatMsgType  MessageType = "CHAT"
)

type Room struct {
	RoomId      string   `json:"room_id"`
	RoomName    string   `json:"room_name"`
	ActiveUsers int      `json:"active_users"`
	Users       []string `json:"users"`
}

type Message struct {
	User    string      `json:"user"`
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}
