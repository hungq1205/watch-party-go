package services

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hungq1205/watch-party/protogen/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (s *RenderService) Start() (*grpc.Server, error) {
	e := echo.New()

	e.Use(middleware.Static("github.com/hungq1205/watch-party/static/"))

	e.GET("/login", LogInPage)
	e.POST("/login", LogIn)
	e.POST("/signup", SignUp)

	err := e.Start(renderServiceAddr)

	return nil, err
}

func LogInPage(c echo.Context) error {
	return c.File("github.com/hungq1205/watch-party/static/views/login.html")
}

func MainPage(c echo.Context) error {
	return c.File("github.com/hungq1205/watch-party/static/views/index.html")
}

func LogIn(c echo.Context) error {
	lock.Lock()
	defer lock.Unlock()

	username := c.FormValue("username")
	password := c.FormValue("password")

	conn := NewGRPCClientConn(userServiceAddr)
	userService := users.NewUserServiceClient(conn)

	res, err := userService.LogIn(c.Request().Context(), &users.LogInRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		if err.Error() == "Incorrect username or password" {
			return c.String(http.StatusBadRequest, err.Error())
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
	userService := users.NewUserServiceClient(conn)

	_, err := userService.SignUp(c.Request().Context(), &users.SignUpRequest{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
	})
	if err != nil {
		if err.Error() == "Username already exists" {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Signed up")
}

func NewGRPCClientConn(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not create client: %v", err)
	}

	return conn
}
