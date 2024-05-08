package internal

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

var MsgBoxes = make(map[int64]*MsgBox)

type Service interface {
	Start() *grpc.Server
}

type Client struct {
	UserId   int64
	Username string
	Conn     *websocket.Conn
	Box      *MsgBox
}

type MsgBox struct {
	Clients []*Client
	OwnerId int64
}

type ClientData struct {
	Datatype   int     `json:"datatype"`
	Username   string  `json:"username"`
	Content    string  `json:"content"`
	BoxUserNum int     `json:"box_user_num"`
	MovieUrl   string  `json:"movie_url"`
	Elapsed    float64 `json:"elapsed"`
	IsPause    bool    `json:"is_pause"`
	IsOwner    bool    `json:"is_owner"`
}

type ClientBoxData struct {
	BoxId   int64 `json:"box_id"`
	IsOwner bool  `json:"is_owner"`
}

type MyCustomClaims struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *MsgBox) Broadcast(fromUserId int64, data *ClientData) error {
	for _, client := range s.Clients {
		if client.UserId == fromUserId && data.Datatype != 2 {
			continue
		}
		data.IsOwner = client.UserId == s.OwnerId
		data.Username = client.Username
		data.BoxUserNum = len(s.Clients)

		err := websocket.JSON.Send(client.Conn, &data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MsgBox) AppendNew(userId int64, username string, conn *websocket.Conn) {
	client := &Client{
		UserId:   userId,
		Username: username,
		Conn:     conn,
		Box:      s,
	}

	for idx, cli := range s.Clients {
		if cli.UserId == client.UserId {
			s.Clients[idx].Conn = conn
			return
		}
	}

	s.Clients = append(s.Clients, client)
}

func AppendNewMsgBox(boxId int64, ownerId int64) {
	MsgBoxes[boxId] = &MsgBox{OwnerId: ownerId}
	MsgBoxes[boxId].Clients = make([]*Client, 0)
}

func (s *MsgBox) Remove(userId int64) {
	for idx, client := range s.Clients {
		if client.UserId == userId {
			s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
			return
		}
	}
}
