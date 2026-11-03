package client

import (
	"chat/message"
	s "chat/server"
	"fmt"
	"net"
)

type Client struct {
	Server   *s.Server
	GroupIds []int
}

func NewClient(ip string, port string) *Client {
	return &Client{
		Server: s.NewServer(ip, port),
	}
}

func (c *Client) Start() {
	con, _ := net.DialUDP("udp", nil, c.Server.Addr)
	defer con.Close()

	go func() {
		for {
			replyRaw := make([]byte, 1024)
			n, _ := con.Read(replyRaw)

			reply, _ := message.Decode(n, replyRaw)

			fmt.Printf("Resposta: %s\n", reply)
		}
	}()

	for {
		var msgTypeRaw uint8
		var group uint16
		var payload []byte

		fmt.Scanf("%d %d %s", &msgTypeRaw, &group, &payload)
		// fmt.Printf("len: %d\n", len(payload))

		msgType, err := message.NewMessageType(msgTypeRaw)
		if err != nil {
			continue
		}

		msg, _ := message.NewMessage(msgType, group, payload)
		// fmt.Printf("len: %d\n", len(msg))

		_, err = con.Write(msg.Encode())
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
