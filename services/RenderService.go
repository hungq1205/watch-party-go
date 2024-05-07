package services

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hungq1205/watch-party/protogen/messages"
	"github.com/hungq1205/watch-party/protogen/movies"
	"github.com/hungq1205/watch-party/protogen/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

var lock = sync.Mutex{}

type RenderService struct {
}

func (s *RenderService) Start() *grpc.Server {
	e := echo.New()

	e.Use(middleware.Static("static"))

	e.GET("/", MainPage)
	e.GET("/login", LogInPage)

	e.POST("/login", LogIn)
	e.POST("/signup", SignUp)
	e.POST("/create", CreateBox)
	e.POST("/delete/:id", DeleteBox)

	err := e.Start(renderServiceAddr)
	if err != nil {
		log.Fatal("Failed to start render server")
	}

	return nil
}

func DeleteBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	boxIdRaw, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	boxId := int64(boxIdRaw)

	conn := NewGRPCClientConn(movieServiceAddr)
	movieServiceClient := movies.NewMovieServiceClient(conn)

	boxRes, err := movieServiceClient.GetBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if boxRes.OwnerId != auth.UserID {
		return c.JSON(http.StatusNotFound, "Can't find target movie box")
	}

	_, err = movieServiceClient.DeleteBox(c.Request().Context(), &movies.MovieBoxIdentifier{
		BoxId: boxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	conn.Close()

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)

	_, err = msgServiceClient.DeleteMessages(c.Request().Context(), &messages.QueryMessageRequest{
		BoxId: boxRes.MsgBoxId,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msgconn.Close()

	return c.NoContent(http.StatusOK)
}

func CreateBox(c echo.Context) error {
	auth, err := Authenticate(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	password := c.QueryParam("password")

	msgconn := NewGRPCClientConn(messageServiceAddr)
	msgServiceClient := messages.NewMessageServiceClient(msgconn)

	msgRes, err := msgServiceClient.CreateMessageBox(c.Request().Context(), &messages.UserGroup{
		UserIds: []int64{auth.UserID},
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msgconn.Close()

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

	return c.JSON(http.StatusOK, res.BoxId)
}

func LogInPage(c echo.Context) error {
	return c.File("static/views/login.html")
}

func MainPage(c echo.Context) error {
	return c.File("static/views/index.html")
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

	return c.String(http.StatusOK, "Logged in")
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

	return c.String(http.StatusCreated, "Signed up")
}

func Authenticate(c echo.Context) (*users.AuthenticateResponse, error) {
	jwtcookie, err := c.Cookie("jwtcookie")
	if err != nil {
		return nil, err
	}

	conn := NewGRPCClientConn(userServiceAddr)
	userServiceClient := users.NewUserServiceClient(conn)
	defer conn.Close()

	return userServiceClient.Authenticate(c.Request().Context(), &users.AuthenticateRequest{JwtToken: jwtcookie.Value})
}

func NewGRPCClientConn(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not create client: %v", err)
	}

	return conn
}
