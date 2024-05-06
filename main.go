package main

import (
	"fmt"

	"github.com/hungq1205/watch-party/services"
	"google.golang.org/grpc"
)

type Service interface {
	Start(c chan *grpc.Server)
}

func main() {
	serviceList := []Service{
		// &services.UserService{},
		&services.MessageService{},
		&services.MovieService{},
	}

	c := make(chan *grpc.Server)

	for idx, service := range serviceList {
		go service.Start(c)
		fmt.Printf("Started service %v...\n", idx+1)
		server := <-c
		defer server.Stop()
	}

	go (&services.RenderService{}).Start()

	fmt.Printf("Started service %v...\n", 3)
	(&services.UserService{}).Start(c)
	server := <-c
	defer server.Stop()
}
