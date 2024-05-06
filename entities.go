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

type msgBox struct {
	Id int64 `json:"id"`
}

type user_chatbox struct {
	UserId int64 `json:"user_id"`
	BoxId  int64 `json:"box_id"`
}

type movie struct {
	MovieId   int64  `json:"movie_id"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	PosterUrl string `json:"poster_url"`
}

type movieBox struct {
	BoxId    int64   `json:"box_id"`
	OwnerId  string  `json:"owner_id"`
	Elapsed  float64 `json:"elapsed"`
	MovieUrl string  `json:"movie_url"`
	Password string  `json:"password"`
}
