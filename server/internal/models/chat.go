package models

type ChatRoom struct {
	RoomId   string `json:"room_id"`
	RoomName string `json:"room_name"`
}

type ChatRoomReqBody struct {
	RoomName string `json:"room_name"`
}

type ChatRoomAddUserReqBody struct {
	RoomId   string   `json:"room_id"`
	RoomName string   `json:"room_name"`
	Users    []string `json:"users"`
}
