package main

type user struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	PwHash      string `json:"pw_hash"`
	DisplayName string `json:"display_name"`
}

type message struct {
	Id       int64  `json:"id"`
	Content  string `json:"content"`
	SenderId int64  `json:"sender_id"`
	BoxId    int64  `json:"box_id"`
}

type chatbox struct {
	Id int64 `json:"id"`
}

type user_chatbox struct {
	UserId int64 `json:"user_id"`
	BoxId  int64 `json:"box_id"`
}
