package models

type Room struct {
	RoomId      string `json:"room_id"`
	RoomName    string `json:"room_name"`
	ActiveUsers int    `json:"active_users"`
}
