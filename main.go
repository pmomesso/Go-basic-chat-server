package main

import (
	"net"
)

func connectionHandler(conn net.Conn) {
	buff := make([]byte, 1)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			panic(err)
		}
		conn.Write(buff[:n])
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
