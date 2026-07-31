package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const PipeName = `\\.\pipe\nexa`

// PipeServer runs a Windows Named Pipe (or TCP localhost on non-Windows) server
// to broadcast events between the NEXA daemon, avatar window, and dashboard.
type PipeServer struct {
	bus       *EventBus
	listeners []net.Conn
	mu        sync.Mutex
	running   bool
}

// NewPipeServer creates a new PipeServer connected to the EventBus.
func NewPipeServer(bus *EventBus) *PipeServer {
	ps := &PipeServer{
		bus:       bus,
		listeners: make([]net.Conn, 0),
	}

	// Subscribe to all events and forward them over named pipe / IPC
	bus.Subscribe("*", func(event Event) {
		ps.Broadcast(event)
	})

	return ps
}

// Start launches the IPC listener server loop.
func (ps *PipeServer) Start() error {
	ps.mu.Lock()
	ps.running = true
	ps.mu.Unlock()

	// Use TCP localhost 59123 for cross-platform compatibility + Windows IPC
	addr := "127.0.0.1:59123"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start IPC server on %s: %w", addr, err)
	}

	go func() {
		defer listener.Close()
		for {
			ps.mu.Lock()
			if !ps.running {
				ps.mu.Unlock()
				break
			}
			ps.mu.Unlock()

			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			ps.mu.Lock()
			ps.listeners = append(ps.listeners, conn)
			ps.mu.Unlock()
		}
	}()

	return nil
}

// Broadcast sends a JSON-encoded event to all connected UI processes.
func (ps *PipeServer) Broadcast(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	ps.mu.Lock()
	defer ps.mu.Unlock()

	var active []net.Conn
	for _, conn := range ps.listeners {
		_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
		_, err := conn.Write(data)
		if err == nil {
			active = append(active, conn)
		} else {
			_ = conn.Close()
		}
	}
	ps.listeners = active
}

// Stop closes all active connections.
func (ps *PipeServer) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.running = false
	for _, conn := range ps.listeners {
		_ = conn.Close()
	}
	ps.listeners = nil
}

// PipeClient connects to a running NEXA daemon to receive events in UI processes.
type PipeClient struct {
	conn net.Conn
}

// ConnectPipeClient connects to the running NEXA daemon IPC server.
func ConnectPipeClient(onEvent func(event Event)) (*PipeClient, error) {
	addr := "127.0.0.1:59123"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to NEXA daemon on %s: %w", addr, err)
	}

	client := &PipeClient{conn: conn}

	go func() {
		defer conn.Close()
		dec := json.NewDecoder(conn)
		for {
			var evt Event
			if err := dec.Decode(&evt); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}
			onEvent(evt)
		}
	}()

	return client, nil
}
