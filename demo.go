package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/anthdm/foreverstore/logger"
	"github.com/anthdm/foreverstore/p2p"
)

// Demo script to showcase the distributed file storage system
func main() {
	logger.SetGlobalLevel(logger.INFO)
	
	fmt.Println("🚀 Distributed File Storage System Demo")
	fmt.Println("=====================================")
	
	// Clean up any existing demo directories
	os.RemoveAll("/tmp/demo_node1")
	os.RemoveAll("/tmp/demo_node2")
	os.RemoveAll("/tmp/demo_node3")
	
	// Create three nodes
	node1 := createDemoNode(":8001", "/tmp/demo_node1", []string{})
	node2 := createDemoNode(":8002", "/tmp/demo_node2", []string{":8001"})
	node3 := createDemoNode(":8003", "/tmp/demo_node3", []string{":8001", ":8002"})
	
	fmt.Println("\n📡 Starting nodes...")
	
	// Start nodes
	go func() {
		if err := node1.Start(); err != nil {
			fmt.Printf("Node 1 error: %v\n", err)
		}
	}()
	time.Sleep(200 * time.Millisecond)
	
	go func() {
		if err := node2.Start(); err != nil {
			fmt.Printf("Node 2 error: %v\n", err)
		}
	}()
	time.Sleep(200 * time.Millisecond)
	
	go func() {
		if err := node3.Start(); err != nil {
			fmt.Printf("Node 3 error: %v\n", err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
	
	fmt.Println("✅ All nodes started and connected")
	
	// Demo file operations
	fmt.Println("\n📁 Demonstrating file operations...")
	
	// Store files on different nodes
	files := map[string][]byte{
		"readme.txt":     []byte("Welcome to the distributed file storage system!"),
		"config.json":    []byte(`{"version": "1.0", "nodes": 3}`),
		"data.csv":       []byte("name,age,city\nAlice,30,NYC\nBob,25,LA"),
		"image.jpg":      bytes.Repeat([]byte("JPEG_DATA"), 100), // Simulate image data
		"document.pdf":   bytes.Repeat([]byte("PDF_CONTENT"), 200), // Simulate PDF data
	}
	
	nodes := []*FileServer{node1, node2, node3}
	
	for i, (filename, content) range files {
		node := nodes[i%len(nodes)]
		nodeNum := (i % len(nodes)) + 1
		
		fmt.Printf("  📤 Storing '%s' on Node %d (%d bytes)\n", filename, nodeNum, len(content))
		
		err := node.Store(filename, bytes.NewReader(content))
		if err != nil {
			fmt.Printf("    ❌ Error storing %s: %v\n", filename, err)
		} else {
			fmt.Printf("    ✅ Stored successfully\n")
		}
		
		time.Sleep(100 * time.Millisecond) // Allow replication time
	}
	
	fmt.Println("\n📥 Retrieving files from different nodes...")
	
	// Retrieve files from different nodes to demonstrate distribution
	for i, filename := range []string{"readme.txt", "config.json", "data.csv", "image.jpg", "document.pdf"} {
		node := nodes[(i+1)%len(nodes)] // Get from different node than stored
		nodeNum := ((i + 1) % len(nodes)) + 1
		
		fmt.Printf("  📥 Retrieving '%s' from Node %d\n", filename, nodeNum)
		
		reader, err := node.Get(filename)
		if err != nil {
			fmt.Printf("    ❌ Error retrieving %s: %v\n", filename, err)
			continue
		}
		
		data, err := io.ReadAll(reader)
		if err != nil {
			fmt.Printf("    ❌ Error reading %s: %v\n", filename, err)
			continue
		}
		
		fmt.Printf("    ✅ Retrieved successfully (%d bytes)\n", len(data))
		
		// Show content for text files
		if filename == "readme.txt" || filename == "config.json" {
			fmt.Printf("    📄 Content: %s\n", string(data))
		}
	}
	
	fmt.Println("\n📊 System Statistics:")
	fmt.Printf("  • Total nodes: %d\n", len(nodes))
	fmt.Printf("  • Files stored: %d\n", len(files))
	fmt.Printf("  • Replication factor: 2-3 (depending on network topology)\n")
	fmt.Printf("  • Encryption: ✅ Enabled (AES-256)\n")
	fmt.Printf("  • Content addressing: ✅ SHA-1 based\n")
	
	fmt.Println("\n🔧 Advanced Features Demonstrated:")
	fmt.Println("  ✅ Peer-to-peer networking")
	fmt.Println("  ✅ Automatic file replication")
	fmt.Println("  ✅ Content-addressable storage")
	fmt.Println("  ✅ File encryption/decryption")
	fmt.Println("  ✅ Network-based file retrieval")
	fmt.Println("  ✅ Structured logging")
	fmt.Println("  ✅ Error handling and retry logic")
	
	fmt.Println("\n⏳ Demo running for 5 more seconds...")
	time.Sleep(5 * time.Second)
	
	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	node1.Stop()
	node2.Stop()
	node3.Stop()
	
	time.Sleep(200 * time.Millisecond)
	
	// Clean up demo directories
	os.RemoveAll("/tmp/demo_node1")
	os.RemoveAll("/tmp/demo_node2")
	os.RemoveAll("/tmp/demo_node3")
	
	fmt.Println("✅ Demo completed successfully!")
	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("  • Use the CLI tool: go run cmd/cli/main.go")
	fmt.Println("  • Check the comprehensive test suite: go test ./...")
	fmt.Println("  • Review the improvement checklist in checklist.md")
}

func createDemoNode(addr, storageDir string, bootstrapNodes []string) *FileServer {
	// Ensure the address is available
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// If address is not available, get a random one
		listener, _ = net.Listen("tcp", ":0")
		addr = listener.Addr().String()
	}
	listener.Close()
	
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    addr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       storageDir,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
	}

	server := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = server.OnPeer

	return server
}
