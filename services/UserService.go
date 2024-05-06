package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hungq1205/watch-party/protogen/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const usr_connectionStr = "root:hungthoi@tcp(127.0.0.1:3306)/user_service"

var usr_lock = sync.Mutex{}

type UserService struct {
	users.UnimplementedUserServiceServer
}

func (s *UserService) Start(c chan *grpc.Server) {
	lis, err := net.Listen("tcp", userServiceAddr)
	if err != nil {
		log.Fatal("Failed to start user service")
	}
	sv := grpc.NewServer()
	c <- sv

	userService := &UserService{}
	users.RegisterUserServiceServer(sv, userService)
	err = sv.Serve(lis)
	if err != nil {
		log.Fatal("Failed to start user service")
	}
}

func (s *UserService) ExistsUsers(ctx context.Context, req *users.ExistsUsersRequest) (*users.ExistsUsersResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	exists := []bool{}
	for _, userId := range req.UserIds {
		row, err := db.Query("SELECT id FROM Users WHERE id=?", userId)
		if err != nil {
			return nil, err
		}

		exists = append(exists, row.Next())
		row.Close()
	}

	return &users.ExistsUsersResponse{Exists: exists}, nil
}

func (s *UserService) SignUp(ctx context.Context, req *users.SignUpRequest) (*users.SignUpResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT id FROM Users WHERE username=?", req.Username)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := h.Sum(nil)

	idRef, err := db.Exec("INSERT INTO Users (username, pw_hash, display_name) VALUES (?, ?, ?)", req.Username, pwHash, req.DisplayName)
	if err != nil {
		return nil, err
	}

	id, err := idRef.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &users.SignUpResponse{UserID: id}, nil
}

func (s *UserService) LogIn(ctx context.Context, req *users.LogInRequest) (*users.LogInResponse, error) {
	fmt.Println("here 2")
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := h.Sum(nil)

	fmt.Println("here 3")
	row, err := db.Query("SELECT id FROM Users WHERE username=? AND pw_hash=?", req.Username, pwHash)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, errors.New("Unable to find user")
	}
	fmt.Println("here 4")

	var id int64
	err = row.Scan(&id)

	tc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(time.Second * 60).Unix(),
	})

	token, err := tc.SignedString([]byte("secret"))
	if err != nil {
		return nil, err
	}

	return &users.LogInResponse{
		UserID:   id,
		JwtToken: token,
	}, nil
}

func (s *UserService) Authenticate(ctx context.Context, req *users.AuthenticateRequest) (*users.AuthenticateResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	token, err := jwt.Parse(req.JwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.InvalidArgument, "unexpected signing method: %v", token)
		}

		return []byte("secret"), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.InvalidArgument, "invalid claims:")
	}

	id := claims["sub"].(int64)
	row, err := db.Query("SELECT username FROM Users WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var username string

	if row.Next() {
		err = row.Scan(&username)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Unable to find user")
	}

	return &users.AuthenticateResponse{
		UserID:   id,
		Username: username,
	}, nil
}

func (s *UserService) GetUsername(ctx context.Context, req *users.GetUsernameRequest) (*users.GetUsernameResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT username FROM Users WHERE id=?", req.UserID)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, errors.New("Unable to find user")
	}

	var username string
	err = row.Scan(&username)

	return &users.GetUsernameResponse{Username: username}, err
}
