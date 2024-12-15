package models

type Room struct {
	RoomId   int    `gorm:"column:room_id;primaryKey;autoIncrement" json:"room_id"`
	RoomName string `gorm:"column:room_name" json:"room_name"`
	Admin    string `gorm:"column:admin" json:"admin"`
}

type UserRooms struct {
	Username string `gorm:"column:username;primaryKey;size:50;not null" json:"username"`
	RoomId   int    `gorm:"column:room_id;primaryKey;not null" json:"room_id"`
}

type CreateRoomReqBody struct {
	RoomName string `json:"room_name"`
}

type AddUsersToRoomReqBody struct {
	RoomId int      `json:"room_id"`
	Users  []string `json:"users"`
}
