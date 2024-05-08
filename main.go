package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hungq1205/watch-party/internal"
	"github.com/hungq1205/watch-party/services"
)

func main() {
	serviceList := []internal.Service{
		&services.UserService{},
		&services.MessageService{},
		&services.MovieService{},
	}

	for _, service := range serviceList {
		sv := service.Start()
		serviceName := strings.Replace(reflect.TypeOf(service).String(), "*services.", "", -1)
		fmt.Printf("Started service %v ...\n", serviceName)
		defer sv.Stop()
	}

	(&services.RenderService{}).Start()
}
