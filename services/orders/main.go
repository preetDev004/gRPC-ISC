package main

func main() {
	gRPCServer := NewGRPCServer(":9000")
	gRPCServer.Run()

	httpServer := NewHTTPServer(":8080")
	httpServer.Run()
}