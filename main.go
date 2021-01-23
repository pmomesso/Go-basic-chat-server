package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const cmdNick = "iam"
const cmdWho = "whoami"
const cmdJoinRoom = "join"
const cmdCreateRoom = "createroom"
const cmdSendMessage = "send"

type server struct {
	availableRooms map[string]*room
	mu             sync.Mutex
}

func newServer() *server {
	retServer := &server{}
	retServer.availableRooms = make(map[string]*room)
	return retServer
}

type room struct {
	mu        sync.Mutex
	name      string
	connected map[string]*client
}

func newRoom(name string) *room {
	retRoom := &room{}
	retRoom.name = name
	retRoom.connected = make(map[string]*client)
	return retRoom
}

type client struct {
	name          string
	connectedRoom *room
	conn          net.Conn
}

func newClient(conn net.Conn) *client {
	ret := &client{name: "unkown", conn: conn}
	return ret
}

func getCommand(received []byte) []string {
	return strings.Split(string(received), " ")
}

func (c *client) sendToNeighbours(message []byte) {
	c.connectedRoom.mu.Lock()
	defer c.connectedRoom.mu.Unlock()
	for _, client := range c.connectedRoom.connected {
		if client != c {
			client.conn.Write(message)
		}
	}
}

func (c *client) handleChangeNickname(newNickname string) {
	oldName := c.name
	c.name = newNickname
	c.conn.Write([]byte(fmt.Sprintf("Ok, I will call you %s\n", c.name)))
	if c.connectedRoom != nil {
		c.sendToNeighbours([]byte(fmt.Sprintf("%s is now %s\n", oldName, c.name)))
	}
}

func (c *client) handleTellNickName() {
	c.conn.Write([]byte(fmt.Sprintf("You are %s\n", c.name)))
}

func (c *client) handleJoinRoom(roomName string, s *server) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.availableRooms[roomName] != nil {
		room := s.availableRooms[roomName]
		if room.connected[c.name] != nil && room.connected[c.name] != c {
			c.conn.Write([]byte(fmt.Sprintf("Sorry, there is already a %s in the room\n", c.name)))
		} else {
			room.mu.Lock()
			room.connected[c.name] = c
			room.mu.Unlock()
			c.connectedRoom = room
			c.sendToNeighbours([]byte(fmt.Sprintf("%s joined the room, say hi!\n", c.name)))
			c.conn.Write([]byte(fmt.Sprintf("Connected to room %s\n", room.name)))
		}
	} else {
		c.conn.Write([]byte(fmt.Sprintf("Sorry, room \"%s\" does not exist\n", roomName)))
	}
}

func (c *client) handleCreateRoom(roomName string, s *server) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.availableRooms[roomName] != nil {
		c.conn.Write([]byte(fmt.Sprintf("Sorry, room \"%s\" already exists\n", roomName)))
	} else {
		room := newRoom(roomName)
		s.availableRooms[roomName] = room
		s.availableRooms[roomName].connected[c.name] = c
		c.connectedRoom = room
		c.conn.Write([]byte(fmt.Sprintf("Created and joined room \"%s\"\n", roomName)))
	}
}

func (c *client) handleSendMessage(messages []string) {
	message := append([]byte(fmt.Sprintf("%s says: ", c.name)+strings.Join(messages, " ")), '\n')
	if c.connectedRoom == nil {
		c.conn.Write([]byte(fmt.Sprintf("Sorry, you have to connect to a room first\n")))
	} else {
		c.sendToNeighbours(message)
	}
}

func (c *client) handleNotUnderstand() {
	c.conn.Write([]byte("Sorry, I do not understand\n"))
}

func connectionHandler(conn net.Conn, s *server) {
	defer conn.Close()
	buff := make([]byte, 50)
	c := newClient(conn)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			return
		}
		command := getCommand(buff[:n-1])
		switch command[0] {
		case cmdNick:
			c.handleChangeNickname(command[1])
		case cmdWho:
			c.handleTellNickName()
		case cmdJoinRoom:
			c.handleJoinRoom(command[1], s)
		case cmdCreateRoom:
			c.handleCreateRoom(command[1], s)
		case cmdSendMessage:
			c.handleSendMessage(command[1:])
		default:
			c.handleNotUnderstand()
		}
	}
}

func main() {
	//Open tcp server
	ln, err := net.Listen("tcp", ":1234") //hardcode port for now
	if err != nil {
		panic(err)
	}
	server := newServer()
	for {
		//Listen for incoming connections
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		//For each incoming connection spawn a go routine to handle it
		go connectionHandler(conn, server)
	}
}
