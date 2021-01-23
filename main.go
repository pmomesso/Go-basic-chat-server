package main

import (
	"fmt"
	"net"
	"strings"
)

const cmdNick = "iam"
const cmdWho = "whoami"

type client struct {
	name string
	conn net.Conn
}

func newClient(conn net.Conn) (ret client) {
	ret = client{name: "unkown", conn: conn}
	return ret
}

func getCommand(received []byte) []string {
	return strings.Split(string(received), " ")
}

func (c *client) handleChangeNickname(newNickname string) {
	c.name = newNickname
	c.conn.Write([]byte(fmt.Sprintf("Ok, I will call you %s\n", c.name)))
}

func (c *client) handleTellNickName(newNickname string) {
	c.conn.Write([]byte(fmt.Sprintf("You are %s\n", c.name)))
}

func (c *client) handleNotUnderstand() {
	c.conn.Write([]byte("Sorry, I do not understand\n"))
}

func connectionHandler(conn net.Conn) {
	buff := make([]byte, 50)
	c := newClient(conn)
	for {
		//Suppose that the entirety of the received bytes constitute a message
		n, err := conn.Read(buff)
		if err != nil {
			panic(err)
		}
		command := getCommand(buff[:n-1])
		switch command[0] {
		case cmdNick:
			c.handleChangeNickname(command[1])
		case cmdWho:
			c.handleTellNickName(command[1])
		default:
			c.handleNotUnderstand()
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
