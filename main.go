package main

import (
	"fmt"
	"log"

	"github.com/hungq1205/watch-party/services"
	"google.golang.org/grpc"
)

func main() {
	c := make(chan *grpc.Server)

	go (&services.UserService{}).Start(c)
	fmt.Println("Started user service...")
	userService := <-c
	defer userService.Stop()

	go (&services.MessageService{}).Start(c)
	fmt.Println("Started message service...")
	msgService := <-c
	defer msgService.Stop()

	go (&services.MovieService{}).Start(c)
	fmt.Println("Started movie service...")
	movieService := <-c
	defer movieService.Stop()

	_, err := (&services.RenderService{}).Start()
	if err != nil {
		log.Fatal("Failed to start render service")
	}
}
