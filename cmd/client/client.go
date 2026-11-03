package main

import "chat/client"

func main() {
	c := client.NewClient("", "5555")
	c.Start()
}
