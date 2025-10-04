package p2p

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	// Get a free port
	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	addr := listener.Addr().String()
	listener.Close()

	opts := TCPTransportOpts{
		ListenAddr:    addr,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, addr)

	// Start listening in a goroutine
	go func() {
		assert.Nil(t, tr.ListenAndAccept())
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Clean up
	assert.Nil(t, tr.Close())
}

func TestTCPTransportDial(t *testing.T) {
	// Create server transport
	serverListener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	serverAddr := serverListener.Addr().String()
	serverListener.Close()

	serverOpts := TCPTransportOpts{
		ListenAddr:    serverAddr,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	serverTr := NewTCPTransport(serverOpts)

	// Start server
	go func() {
		serverTr.ListenAndAccept()
	}()
	time.Sleep(100 * time.Millisecond)

	// Create client transport
	clientListener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	clientAddr := clientListener.Addr().String()
	clientListener.Close()

	clientOpts := TCPTransportOpts{
		ListenAddr:    clientAddr,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	clientTr := NewTCPTransport(clientOpts)

	// Test dial
	err = clientTr.Dial(serverAddr)
	assert.Nil(t, err)

	// Clean up
	time.Sleep(100 * time.Millisecond)
	serverTr.Close()
	// Don't close clientTr as it doesn't have a listener
}
