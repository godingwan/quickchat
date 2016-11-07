package main

import (
	"fmt"
	"net"
	"time"
)

var cns []net.Conn

var msg chan string

type client struct {
	conn     *net.Conn
	username string
}

func main() {
	fmt.Println("Starting quickchat...")
	msg = make(chan string)
	netInit(10)
	go func() {
		for m := range msg {
			fmt.Println(m)
		}
	}()
	for {
		time.Sleep(3 * time.Second)
		// send test string to all connected clients
		for c := range cns {
			fmt.Fprint(cns[c], "Hello fellow client!\n")
		}
	}
}

func netInit(connectionLimit int) {
	cns = make([]net.Conn, 0, connectionLimit)
	// start broadcasting for availability using Dial()
	requestConnections()
	// start listening for requests using Listen()
	go acceptConnections()
	// start listening for responses using Listen()
	go handshake()
}

// broadcast availability on port 875
func requestConnections() {
	fmt.Println("Sending UDP broadcast")
	conn, err := net.Dial("udp", "255.255.255.255:875")
	if err != nil {
		fmt.Printf("Broadcast error: %v\n", err)
		return
	}
	conn.Write([]byte("request tcp"))
	conn.Close()
}

// listen for tcp responses on 876
func acceptConnections() {
	laddr := net.TCPAddr{
		IP:   nil,
		Port: 876,
	}
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		fmt.Printf("Can't accept connections! Err: %v\n", err)
	}
	for {
		conn, err := ln.AcceptTCP()
		fmt.Println("Accepted tcp connection.")
		if err != nil {
			fmt.Printf("Error accepting connections: %v\n", err)
		} else {
			fmt.Printf("Accepted tcp connection from %s\n", conn.RemoteAddr())
			cns = append(cns, conn)
			readIncomingMessages(conn)
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
	fmt.Println("Listening for UDP...")
	ser, err := net.ListenUDP("udp", &udpAddr)
	if err != nil {
		fmt.Printf("Error responding to UDP broadcast: %v\n", err)
	}
	for {
		fmt.Println("Accepting UDP...")
		_, remoteAddr, err := ser.ReadFromUDP(buff)
		if err == nil {
			fmt.Printf("Recieved msg '%s'\nfrom addr: %v\n", buff, remoteAddr)
			go makeConnection(remoteAddr)
		}
	}
}

func makeConnection(remoteAddr *net.UDPAddr) {
	toDial := net.TCPAddr{
		Port: 876,
		IP:   remoteAddr.IP,
	}
	fmt.Printf("Making connection with %s\n", toDial.IP.String())
	conn, err := net.DialTCP("tcp", nil, &toDial)
	if err == nil {
		fmt.Printf("Successful connection with %s\n", remoteAddr.String())
		cns = append(cns, conn)
		readIncomingMessages(conn)
	}
}

func readIncomingMessages(c *net.TCPConn) {
	buff := make([]byte, 2048)
	for {
		n, err := c.Read(buff)
		if err != nil {
			fmt.Printf("TCP read error, closing connection: %v\n", err)
			c.Close()
			return
		}
		msg <- string(buff[:n])
	}
}
