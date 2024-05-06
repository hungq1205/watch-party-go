package services

import (
	"context"
	"database/sql"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	mes "github.com/hungq1205/watch-party/messages"
	"github.com/hungq1205/watch-party/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MessageService struct {
	mes.UnimplementedMessageServiceServer
}

const msg_connectionStr = "root:hungthoi@tcp(127.0.0.1:3307)/message_service"

var msg_lock = sync.Mutex{}

func (s *MessageService) RemoveUserFromBox(context.Context, *MessageBoxIdentifier) (*ActionResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Exec("DELETE FROM MsgBox_user WHERE user_id=?", req.UserId)
	if err != nil {
		return nil, err
	}

	aff, err := row.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &ActionResponse{Success: aff > 0}, nil
}

func (s *MessageService) AddUserToBox(context.Context, *AddUserToBoxRequest) (*ActionResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Exec("INSERT INTO MsgBox_user (box_id, user_id) VALUES (?, ?)", req.BoxId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &ActionResponse{Success: aff > 0}, nil
}

func (s *MessageService) CreateMessageBox(ctx context.Context, req *mes.UserGroup) (*mes.MessageBoxIdentifier, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	idRef, err := db.Exec("INSERT INTO MessageBoxes VALUES (NULL)")
	if err != nil {
		return nil, err
	}

	box_id, err := idRef.LastInsertId()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(":3001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not create client: %v", err)
	}
	defer conn.Close()

	userClient := users.NewUserServiceClient(conn)
	exists, err := userClient.ExistsUsers(ctx, &users.ExistsUsersRequest{UserIds: req.UserIds})
	if err != nil {
		return nil, err
	}

	for idx, exist := range exists.Exists {
		if exist {
			_, err := db.Exec("INSERT INTO MsgBox_User (box_id, user_id) VALUES (?, ?)", box_id, req.UserIds[idx])
			if err != nil {
				return nil, err
			}
		}
	}

	return &mes.MessageBoxIdentifier{BoxId: box_id}, nil
}

func (s *MessageService) DeleteMessageBox(ctx context.Context, req *mes.MessageBoxIdentifier) (*mes.ActionResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Exec("DELETE FROM MessageBoxes WHERE box_id=?", req.BoxId)
	if err != nil {
		return nil, err
	}

	aff, err := row.RowsAffected()
	if err != nil {
		return nil, err
	}

	if aff == 0 {
		return &mes.ActionResponse{Success: false}, nil
	}

	row, err = db.Exec("DELETE FROM MsgBox_User WHERE box_id=?", req.BoxId)
	if err != nil {
		return nil, err
	}

	_, err = s.DeleteMessages(ctx, &mes.QueryMessageRequest{
		BoxId:  req.BoxId,
		UserId: -1,
	})

	if err != nil {
		return nil, err
	}

	return &mes.ActionResponse{Success: true}, nil
}

func (s *MessageService) GetMessageBox(ctx context.Context, req *mes.MessageBoxIdentifier) (*mes.UserGroup, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT user_id FROM MsgBox_User WHERE box_id=?", req.BoxId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	userIds := []int64{}
	for row.Next() {
		var userId int64
		err = row.Scan(&userId)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return &mes.UserGroup{UserIds: userIds}, nil
}

func (s *MessageService) GetMessageBoxOfUser(ctx context.Context, req *mes.UserIdentifier) (*mes.BoxGroup, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT box_id FROM MsgBox_User WHERE user_id=?", req.UserId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	boxIds := []int64{}
	for row.Next() {
		var boxId int64
		err = row.Scan(&boxId)
		if err != nil {
			return nil, err
		}
		boxIds = append(boxIds, boxId)
	}

	return &mes.BoxGroup{BoxIds: boxIds}, nil
}

func (s *MessageService) Sent(ctx context.Context, req *mes.SentRequest) (*mes.MessageIdentifier, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	idRef, err := db.Exec("INSERT INTO Messages (user_id, box_id, content) VALUES (?, ?, ?)", req.UserId, req.BoxId, req.Content)
	if err != nil {
		return nil, err
	}

	id, err := idRef.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &mes.MessageIdentifier{MessageId: id}, nil
}

func (s *MessageService) Delete(ctx context.Context, req *mes.MessageIdentifier) (*mes.ActionResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Exec("DELETE FROM Messages WHERE message_id=?", req.MessageId)
	if err != nil {
		return nil, err
	}

	aff, err := row.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &mes.ActionResponse{Success: aff > 0}, nil
}

func (s *MessageService) GetMessage(ctx context.Context, req *mes.MessageIdentifier) (*mes.Message, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT * FROM Messages WHERE message_id=?", req.MessageId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, err
	}

	var m mes.Message
	err = row.Scan(&m.MessageId, &m.UserId, &m.BoxId, &m.Content)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *MessageService) QueryMessages(ctx context.Context, req *mes.QueryMessageRequest) (*mes.QueryMessageResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var row *sql.Rows
	if req.BoxId == -1 {
		row, err = db.Query("SELECT * FROM Messages WHERE user_id=?", req.UserId)
	} else if req.UserId == -1 {
		row, err = db.Query("SELECT * FROM Messages WHERE box_id=?", req.BoxId)
	} else {
		row, err = db.Query("SELECT * FROM Messages WHERE box_id=? AND user_id=?", req.BoxId, req.UserId)
	}

	if err != nil {
		return nil, err
	}
	defer row.Close()

	messages := []*mes.Message{}
	for row.Next() {
		var m mes.Message
		err = row.Scan(&m.MessageId, &m.UserId, &m.BoxId, &m.Content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}

	return &mes.QueryMessageResponse{Messages: messages}, nil
}

func (s *MessageService) DeleteMessages(ctx context.Context, req *mes.QueryMessageRequest) (*mes.ActionResponse, error) {
	msg_lock.Lock()
	defer msg_lock.Unlock()

	db, err := sql.Open("mysql", msg_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rslt sql.Result
	if req.BoxId == -1 {
		rslt, err = db.Exec("DELETE FROM Messages WHERE user_id=?", req.UserId)
	} else if req.UserId == -1 {
		rslt, err = db.Exec("DELETE FROM Messages WHERE box_id=?", req.BoxId)
	} else {
		rslt, err = db.Exec("DELETE FROM Messages WHERE box_id=? AND user_id=?", req.BoxId, req.UserId)
	}
	if err != nil {
		return nil, err
	}

	eff, err := rslt.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &mes.ActionResponse{Success: eff > 0}, nil
}
