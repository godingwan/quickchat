package main

import (
	"fmt"
	"net"
	"time"
)

var cns []net.Conn

func main() {
	fmt.Println("Starting quickchat...")
	netInit(10)
	time.Sleep(10 * time.Second)
}

func netInit(connectionLimit int) {
	cns = make([]net.Conn, 0, connectionLimit)
	// start broadcasting for availability using Dial()
	go requestConnections(2000)
	// start listening for requests using Listen()
	go acceptConnections()
	// start listening for responses using Listen()
	go handshake()
}

// repeatedly broadcast availability on port 875
func requestConnections(timeout int) {
	for {
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		conn, err := net.Dial("udp", "255.255.255.255:875")
		if err != nil {
			fmt.Printf("Some error: %v\n", err)
			continue
		}
		fmt.Fprintf(conn, "request tcp")
	}
}

// listen for tcp responses on 876
func acceptConnections() {
	ln, err := net.Listen("tcp", ":876")
	if err != nil {
		fmt.Printf("Can't accept connections! Err: %v\n", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error accepting connections: %v\n", err)
		} else {
			fmt.Printf("Accepted tcp connection from %s\n", conn.RemoteAddr())
			cns = append(cns, conn)
		}
	}
}

// listen for broadcasts on 875 and send tcp connection requests on 876
func handshake() {
	// buffer
	buff := make([]byte, 2048)
	// listen for udp
	udpAddr := net.UDPAddr{
		Port: 875,
		IP:   net.ParseIP("255.255.255.255"),
	}
	ser, err := net.ListenUDP("udp", &udpAddr)
	if err != nil {
		fmt.Printf("Error responding to UDP broadcast: %v\n", err)
	}
	for {
		_, remoteAddr, err := ser.ReadFromUDP(buff)
		if err == nil {
			fmt.Printf("Recieved msg '%s'\nfrom addr: %v\n", buff, remoteAddr)
		}
	}
}
