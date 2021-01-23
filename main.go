package main

import (
	"fmt"
	"net"
	"strings"
)

const CMD_NICK = "iam"
const CMD_WHO = "whoami"

type client struct {
	name string
}

func newClient() (ret client) {
	ret = client{name: "unkown"}
	return ret
}

func connectionHandler(conn net.Conn) {
	buff := make([]byte, 50)
	c := newClient()
	for {
		//Suppose that the entirety of the received bytes constitute a message
		n, err := conn.Read(buff)
		if err != nil {
			panic(err)
		}
		command := strings.Split(string(buff[:n-1]), " ")
		switch command[0] {
		case CMD_NICK:
			c.name = command[1]
			conn.Write([]byte(fmt.Sprintf("Ok, i will call you %s from now on\n", c.name)))
		case CMD_WHO:
			conn.Write([]byte(fmt.Sprintf("You are %s\n", c.name)))
		default:
			conn.Write([]byte(fmt.Sprintf("I am sorry, I do not understand\n")))
		}
	}
}

func main() {
	//Open tcp server
	server, err := net.Listen("tcp", ":1234") //hardcode port for now
	if err != nil {
		panic(err)
	}
	for {
		//Listen for incoming connections
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		//For each incoming connection spawn a go routine to handle it
		go connectionHandler(conn)
	}
}
