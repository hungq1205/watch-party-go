package services

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/hungq1205/watch-party/protogen/movies"
	"github.com/hungq1205/watch-party/protogen/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const mv_connectionStr = "root:hungthoi@tcp(127.0.0.1:3308)/movie_service"

var mv_lock = sync.Mutex{}

type MovieService struct {
}

func (s *MovieService) CreateBox(ctx context.Context, req *movies.CreateRequest) (*movies.MovieBoxIdentifier, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	conn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	userClient := users.NewUserServiceClient(conn)
	exists, err := userClient.ExistsUsers(ctx, &users.ExistsUsersRequest{UserIds: []int64{req.OwnerId}})
	if err != nil {
		return nil, err
	}

	if !exists.Exists[0] {
		return nil, errors.New("Movie box owner doesn't exists")
	}

	idRef, err := db.Exec("INSERT INTO MovieBoxes (owner_id, password) VALUES (?, ?)", req.OwnerId, req.Password)
	if err != nil {
		return nil, err
	}

	box_id, err := idRef.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &movies.MovieBoxIdentifier{BoxId: box_id}, nil
}

func (s *MovieService) AddToBox(ctx context.Context, req *movies.UserBoxRequest) (*movies.ActionResponse, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO MvBox_User (box_id, user_id) VALUES (?, ?)", req.BoxId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &movies.ActionResponse{Success: true}, nil
}

func (s *MovieService) RemoveFromBox(ctx context.Context, req *movies.UserBoxRequest) (*movies.ActionResponse, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Exec("DELETE FROM MvBox_user WHERE user_id=?", req.UserId)
	if err != nil {
		return nil, err
	}

	aff, err := row.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &movies.ActionResponse{Success: aff > 0}, nil
}

func (s *MovieService) GetBox(ctx context.Context, req *movies.MovieBoxIdentifier) (*movies.MovieBox, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT * FROM MovieBoxes WHERE box_id=?", req.BoxId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil
	}

	var box movies.MovieBox
	err = row.Scan(&box.BoxId, &box.OwnerId, &box.Elapsed, &box.MovieUrl, &box.Password)
	if err != nil {
		return nil, err
	}

	return &box, nil
}

func (s *MovieService) SetBox(ctx context.Context, req *movies.MovieBox) (*movies.ActionResponse, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE MovieBoxes SET owner_id=?, elapsed=?, movie_url=?, password=? WHERE box_id=?", req.OwnerId, req.Elapsed, req.MovieUrl, req.Password, req.BoxId)
	if err != nil {
		return nil, err
	}

	return &movies.ActionResponse{Success: true}, nil
}

func (s *MovieService) UserOfBox(ctx context.Context, req *movies.MovieBoxIdentifier) (*movies.UserGroup, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT user_id FROM MvBox_User WHERE box_id=?", req.BoxId)
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

	return &movies.UserGroup{UserIds: userIds}, nil
}

func (s *MovieService) GetMovie(ctx context.Context, req *movies.MovieIdentifier) (*movies.Movie, error) {
	mv_lock.Lock()
	defer mv_lock.Unlock()

	db, err := sql.Open("mysql", mv_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT * FROM Movies WHERE movie_id=?", req.MovieId)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil
	}

	var mv movies.Movie
	err = row.Scan(&mv.MovieId, &mv.Title, &mv.Url, &mv.PosterUrl)
	if err != nil {
		return nil, err
	}

	return &mv, nil
}
