package main

import "chat/server"

func main() {
	s := server.NewServer("", "5555")
	s.Listen()
}
