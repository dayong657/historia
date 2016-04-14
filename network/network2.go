package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	CONN_TYPE = "tcp"
)

var (
	NotEnoughAliveHosts = errors.New("There are not enough alive hosts to fulfill your request")
)

type Node string

type NetworkPacket struct {
	Source             Node
	DestinationHandler int64
	Data               []byte
}

type RequestHandler func(packet *NetworkPacket)

func NewNetwork(id int) *Network {
	n := &Network{
		hosts:          []string{"localhost:3333", "localhost:4444", "localhost:5555"},
		id:             id,
		requestHandler: make(map[int64]RequestHandler),
	}

	go n.start()
	return n
}

type Network struct {
	hosts          []string
	id             int
	requestHandler map[int64]RequestHandler
	handlerLock    sync.RWMutex
}

func (n *Network) RegisterHandler(handler RequestHandler, timeout time.Duration) int64 {
	port := rand.Int63()
	n.RegisterKnownHandler(port, handler)

	return port
}

func (n *Network) RegisterKnownHandler(port int64, handler RequestHandler) {
	n.handlerLock.Lock()
	defer n.handlerLock.Unlock()

	n.requestHandler[port] = handler
}

func (n *Network) UnregisterHandler(port int64) {
	n.handlerLock.Lock()
	defer n.handlerLock.Unlock()

	delete(n.requestHandler, port)
}

//myAddress gets the address of this network item
func (n *Network) MyAddress() Node {
	return Node(n.hosts[n.id])
}

func (n *Network) GetNodes() []Node {
	var nodes []Node
	for _, host := range n.hosts {
		nodes = append(nodes, Node(host))
	}
	return nodes
}

func (n *Network) start() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, string(n.MyAddress()))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + string(n.MyAddress()))
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go n.handleRequest(conn)
	}
}

func (n *Network) Send(address Node, destinationPort int64, message []byte) error {

	nm := NetworkPacket{DestinationHandler: destinationPort, Data: message}
	return n.send(address, nm)
}

func (n *Network) send(address Node, message NetworkPacket) error {
	message.Source = n.MyAddress()

	conn, err := net.Dial(CONN_TYPE, string(address))
	if err != nil {
		return err
	}

	defer conn.Close()

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	conn.Write(b)

	return nil
}

func (n *Network) handleRequest(conn net.Conn) {
	defer conn.Close()

	// TODO use a buffer pool or something for this in the future.
	// the protocol assumes a non-byzantine model though so it's fine
	// until we get too big of messages
	incomingBuffer := make([]byte, 1024*1024*4)

	dataLength, err := conn.Read(incomingBuffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	// truncate buffer to what was actually read
	resizedBuffer := incomingBuffer[:dataLength]

	var packet NetworkPacket
	err = json.Unmarshal(resizedBuffer, &packet)

	if err != nil {
		fmt.Println("Error decoding:", err.Error())
		fmt.Println(string(resizedBuffer))

		return
	}

	n.handlerLock.RLock()
	handler, ok := n.requestHandler[packet.DestinationHandler]
	n.handlerLock.RUnlock()

	if ok {
		handler(&packet)
	} else {
		fmt.Printf("error, unknown message destination: %d\n", packet.DestinationHandler)
	}
}
