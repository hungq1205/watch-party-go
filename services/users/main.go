package main

func main() {
	grpcServer := NewGRPCServer(":3000")
	grpcServer.Run()
}
