package main

func main() {
	server := NewAPIServer(":4000")
	server.Run()
}
