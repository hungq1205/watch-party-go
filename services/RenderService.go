package services

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hungq1205/watch-party/internal"
	"github.com/hungq1205/watch-party/protogen/messages"
	"github.com/hungq1205/watch-party/protogen/movies"
	"github.com/hungq1205/watch-party/protogen/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	renderServiceAddr  = ":3000"
	userServiceAddr    = ":3001"
	messageServiceAddr = ":3002"
	movieServiceAddr   = ":3003"
)

var (
	lock = sync.Mutex{}
)

type RenderService struct {
}

func (s *RenderService) Start() *grpc.Server {
	e := echo.New()

	e.Use(middleware.Static("static"))

	e.GET("/", MainPage)
	e.GET("/box", MainPage)
	e.GET("/login", LogInPage)
	e.GET("/lobby", LobbyPage)
	e.GET("/clientBoxData", ClientBoxData)
	e.GET("/ws", MessageHandler)

	e.POST("/join", JoinBox)
	e.POST("/login", LogIn)
	e.POST("/signup", SignUp)
	e.POST("/create", CreateBox)
	e.POST("/delete", DeleteBox)
	e.POST("/leave", LeaveBox)

	err := e.Start(renderServiceAddr)
	if err != nil {
		log.Fatal("Failed to start render server")
	}

	return nil
}

func LogInPage(c echo.Context) error {
	_, err := Authenticate(c)
	if err == nil {
		return c.Redirect(http.StatusPermanentRedirect, "/lobby")
	}

	return c.File("static/views/login.html")
}

func LobbyPage(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	defer conn.Close()

	_, err = movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{
		UserId: auth.UserID,
	})
	if err == nil {
		return c.Redirect(http.StatusMovedPermanently, "/box")
	} else if status.Code(err) == codes.NotFound {
		err = nil
	}

	return c.File("static/views/lobby.html")
}

func MainPage(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.Redirect(http.StatusPermanentRedirect, "/login")
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	defer conn.Close()

	_, err = movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{
		UserId: auth.UserID,
	})
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	return c.File("static/views/index.html")
}

func MessageHandler(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	box, err := movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{
		UserId: auth.UserID,
	})
	if err != nil {
		conn.Close()
		return c.String(http.StatusNotFound, err.Error())
	}
	mvBox, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: box.BoxId,
	})
	if err != nil {
		conn.Close()
		return c.String(http.StatusNotFound, err.Error())
	}
	conn.Close()

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		if internal.MsgBoxes[mvBox.MsgBoxId].Clients == nil {
			internal.AppendNewMsgBox(mvBox.MsgBoxId, auth.UserID)
		}
		internal.MsgBoxes[mvBox.MsgBoxId].AppendNew(auth.UserID, auth.Username, ws)
		for {
			var data internal.ClientData
			err = websocket.JSON.Receive(ws, &data)
			if err != nil {
				break
			}

			err = internal.MsgBoxes[mvBox.MsgBoxId].Broadcast(auth.UserID, &data)
			if err != nil {
				break
			}
		}
		if auth.UserID == mvBox.OwnerId {
			c.Logger().Print(fmt.Sprintf("deleted box %v", box.BoxId))
			err = UncheckDeleteBox(c, box.BoxId)
			if err != nil && status.Code(err) == codes.NotFound {
				err = nil
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return err
}

func JoinBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	password := c.FormValue("password")
	rawBoxId, err := strconv.Atoi(c.FormValue("box_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	boxId := int64(rawBoxId)

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	defer conn.Close()

	_, err = movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{
		UserId: auth.UserID,
	})
	if err == nil {
		return c.String(http.StatusConflict, err.Error())
	} else if status.Code(err) != codes.NotFound {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	boxRes, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxId,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "Movie box doesn't exist")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if boxRes.Password != password {
		return c.String(http.StatusBadRequest, "Incorrect password or box ID for movie box")
	}

	_, err = movieServiceClient.AddToBox(c.Request().Context(), &movies.UserBoxRequest{
		UserId: auth.UserID,
		BoxId:  boxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)
	defer msgconn.Close()

	_, err = msgServiceClient.AddUserToBox(c.Request().Context(), &messages.UserBox{
		UserId: auth.UserID,
		BoxId:  boxRes.MsgBoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func DeleteBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)

	boxIdRes, err := movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{UserId: auth.UserID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		return c.String(http.StatusBadRequest, err.Error())
	}

	boxRes, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxIdRes.BoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if boxRes.OwnerId != auth.UserID {
		return c.String(http.StatusNotFound, "Can't find target movie box")
	}

	_, err = movieServiceClient.DeleteBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxIdRes.BoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	conn.Close()

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)

	_, err = msgServiceClient.DeleteMessageBox(c.Request().Context(), &messages.MessageBoxIdentifier{
		BoxId: boxRes.MsgBoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msgconn.Close()

	delete(internal.MsgBoxes, boxRes.MsgBoxId)

	return c.NoContent(http.StatusOK)
}

func LeaveBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)

	boxIdRes, err := movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{UserId: auth.UserID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		return c.String(http.StatusBadRequest, err.Error())
	}

	boxRes, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxIdRes.BoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if boxRes.OwnerId == auth.UserID {
		return c.String(http.StatusBadRequest, "You are the owner")
	}

	_, err = movieServiceClient.RemoveFromBox(c.Request().Context(), &movies.UserBoxRequest{
		BoxId:  boxIdRes.BoxId,
		UserId: auth.UserID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	conn.Close()

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)

	_, err = msgServiceClient.RemoveUserFromBox(c.Request().Context(), &messages.UserBox{
		BoxId:  boxRes.MsgBoxId,
		UserId: auth.UserID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msgconn.Close()

	internal.MsgBoxes[boxRes.MsgBoxId].Remove(auth.UserID)

	return c.NoContent(http.StatusOK)
}

func UncheckDeleteBox(c echo.Context, boxId int64) error {
	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)

	boxRes, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxId,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Error(codes.NotFound, err.Error())
		}

		return err
	}

	_, err = movieServiceClient.DeleteBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxId,
	})
	if err != nil {
		return err
	}
	conn.Close()

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)

	_, err = msgServiceClient.DeleteMessageBox(c.Request().Context(), &messages.MessageBoxIdentifier{
		BoxId: boxRes.MsgBoxId,
	})
	if err != nil {
		return err
	}
	msgconn.Close()

	delete(internal.MsgBoxes, boxRes.MsgBoxId)

	return nil
}

func CreateBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	password := c.FormValue("password")

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)
	defer msgconn.Close()

	msgRes, err := msgServiceClient.CreateMessageBox(c.Request().Context(), &messages.UserGroup{
		UserIds: []int64{},
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	defer conn.Close()

	res, err := movieServiceClient.CreateBox(c.Request().Context(), &movies.CreateRequest{
		OwnerId:  auth.UserID,
		MsgBoxId: msgRes.BoxId,
		Password: password,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	internal.AppendNewMsgBox(msgRes.BoxId, auth.UserID)

	return c.JSON(http.StatusOK, res)
}

func LogIn(c echo.Context) error {
	lock.Lock()
	defer lock.Unlock()

	username := c.FormValue("username")
	password := c.FormValue("password")

	conn := NewGRPCClientConn(userServiceAddr)
	userServiceClient := users.NewUserServiceClient(conn)
	defer conn.Close()

	res, err := userServiceClient.LogIn(c.Request().Context(), &users.LogInRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusBadRequest, "Incorrect username or password")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	token := &http.Cookie{
		Name:    "jwtcookie",
		Value:   res.JwtToken,
		Expires: time.Now().Add(time.Minute * 30),
	}
	c.SetCookie(token)

	return c.NoContent(http.StatusOK)
}

func SignUp(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	displayName := c.FormValue("display_name")

	conn := NewGRPCClientConn(userServiceAddr)
	userServiceClient := users.NewUserServiceClient(conn)
	defer conn.Close()

	_, err := userServiceClient.SignUp(c.Request().Context(), &users.SignUpRequest{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return c.String(http.StatusBadRequest, "Username already exists")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func ClientBoxData(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusUnauthorized, err.Error())
	}

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)
	defer conn.Close()

	boxId, err := movieServiceClient.BoxOfUser(c.Request().Context(), &movies.BoxOfUserRequest{
		UserId: auth.UserID,
	})
	if err != nil {
		return c.Redirect(http.StatusMovedPermanently, "/lobby")
	}

	box, err := movieServiceClient.GetBox(c.Request().Context(), boxId)
	if err != nil {
		return c.Redirect(http.StatusMovedPermanently, "/lobby")
	}

	data := &internal.ClientBoxData{
		BoxId:   boxId.BoxId,
		IsOwner: auth.UserID == internal.MsgBoxes[box.MsgBoxId].OwnerId,
	}

	return c.JSON(http.StatusOK, data)
}

func Authenticate(c echo.Context) (*users.AuthenticateResponse, error) {
	jwtcookie, err := c.Cookie("jwtcookie")
	if err != nil || jwtcookie == nil {
		c.Logger().Printf("unauth")
		return nil, status.Error(codes.Unauthenticated, "Unauthenticate")
	}

	conn := NewGRPCClientConn(userServiceAddr)
	userServiceClient := users.NewUserServiceClient(conn)
	defer conn.Close()

	value := jwtcookie.Value
	req := users.AuthenticateRequest{JwtToken: value}
	res, err := userServiceClient.Authenticate(c.Request().Context(), &req)

	return res, err
}

func NewGRPCClientConn(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not create client: %v", err)
	}

	return conn
}
