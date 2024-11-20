package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"
)

type Metrics struct {
	Cpu    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

var clients = make(map[net.Conn]bool)
var mu sync.Mutex

func handleConnection(c net.Conn) {
	defer func() {
		mu.Lock()
		delete(clients, c)
		mu.Unlock()
		c.Close()
	}()

	mu.Lock()
	clients[c] = true
	mu.Unlock()

	buffer := make([]byte, 1024)

	defer c.Close()

	for {
		n, err := c.Read(buffer)

		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnected, %s\n", c.RemoteAddr().String())
				return
			}
			log.Printf("Error reading data, %v\n", err)
			return
		}

		var metrics Metrics

		if err := json.Unmarshal(buffer[:n], &metrics); err != nil {
			log.Printf("Invalid data from %s: %v\n", c.RemoteAddr().String(), err)
			continue
		}

		log.Printf("Metrics from %s - CPU: %.2f%%, Memory: %.2fGB\n", c.RemoteAddr().String(), metrics.Cpu, metrics.Memory)
	}
}

func main() {
	l, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("Something went wrong, %v\n", err)
	}

	log.Printf("Server is starting on port 8080...")

	defer l.Close()

	for {
		c, err := l.Accept()
		log.Printf("Serving %s\n", c.RemoteAddr().String())

		if err != nil {
			log.Printf("Connection error, %v\n", err)
			return
		}

		go handleConnection(c)
	}
}
