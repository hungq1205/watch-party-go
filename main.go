package main

import (
	"log"
	"net"

	"github.com/hungq1205/watch-party/services"
	"github.com/hungq1205/watch-party/users"
	"google.golang.org/grpc"
)

const userServicePort = ":3001"
const messageServicePort = ":3002"

func main() {
	lis, err := net.Listen("tcp", userServicePort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	defer s.Stop()

	userService := &services.UserService{}
	users.RegisterUserServiceServer(s, userService)
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
