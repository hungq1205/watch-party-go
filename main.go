package main

import (
	"fmt"
	"log"

	"github.com/hungq1205/watch-party/services"
)

func main() {
	// userService, err := (&services.UserService{}).Start()
	// if err != nil {
	// 	log.Fatal("Failed to start user service")
	// }
	// defer userService.Stop()
	// fmt.Print("Started user service")

	// msgService, err := (&services.MessageService{}).Start()
	// if err != nil {
	// 	log.Fatal("Failed to start message service")
	// }
	// defer msgService.Stop()
	// fmt.Print("Started message service")

	// movieService, err := (&services.MovieService{}).Start()
	// if err != nil {
	// 	log.Fatal("Failed to start movie service")
	// }
	// defer movieService.Stop()
	// fmt.Print("Started movie service")

	_, err := (&services.RenderService{}).Start()
	if err != nil {
		log.Fatal("Failed to start render service")
	}
	fmt.Print("Started render service")
}
