package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/anthdm/foreverstore/p2p"
	"github.com/stretchr/testify/assert"
)

func TestFileServerIntegration(t *testing.T) {
	// Create temporary directories for testing
	tempDir1 := "/tmp/fs_test_1"
	tempDir2 := "/tmp/fs_test_2"
	defer func() {
		os.RemoveAll(tempDir1)
		os.RemoveAll(tempDir2)
	}()

	// Get free ports
	listener1, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	addr1 := listener1.Addr().String()
	listener1.Close()

	listener2, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	addr2 := listener2.Addr().String()
	listener2.Close()

	// Create first server
	server1 := createTestServer(addr1, tempDir1, []string{})
	
	// Create second server that connects to first
	server2 := createTestServer(addr2, tempDir2, []string{addr1})

	// Start servers
	go func() {
		err := server1.Start()
		if err != nil {
			t.Logf("Server 1 error: %v", err)
		}
	}()

	// Give first server time to start
	time.Sleep(100 * time.Millisecond)

	go func() {
		err := server2.Start()
		if err != nil {
			t.Logf("Server 2 error: %v", err)
		}
	}()

	// Give servers time to connect
	time.Sleep(500 * time.Millisecond)

	// Test file storage and retrieval
	testData := []byte("Hello, distributed file system!")
	testKey := "test_file.txt"

	// Store file on server1
	err = server1.Store(testKey, bytes.NewReader(testData))
	assert.Nil(t, err)

	// Give time for replication
	time.Sleep(200 * time.Millisecond)

	// Retrieve file from server1
	reader1, err := server1.Get(testKey)
	assert.Nil(t, err)

	data1, err := io.ReadAll(reader1)
	assert.Nil(t, err)
	assert.Equal(t, testData, data1)

	// Retrieve file from server2 (should get it via network)
	reader2, err := server2.Get(testKey)
	assert.Nil(t, err)

	data2, err := io.ReadAll(reader2)
	assert.Nil(t, err)
	assert.Equal(t, testData, data2)

	// Clean up
	server1.Stop()
	server2.Stop()
	time.Sleep(100 * time.Millisecond)
}

func TestFileServerReplication(t *testing.T) {
	// Create temporary directories
	tempDirs := []string{"/tmp/fs_repl_1", "/tmp/fs_repl_2", "/tmp/fs_repl_3"}
	defer func() {
		for _, dir := range tempDirs {
			os.RemoveAll(dir)
		}
	}()

	// Get free ports
	var addrs []string
	for i := 0; i < 3; i++ {
		listener, err := net.Listen("tcp", ":0")
		assert.Nil(t, err)
		addrs = append(addrs, listener.Addr().String())
		listener.Close()
	}

	// Create servers
	server1 := createTestServer(addrs[0], tempDirs[0], []string{})
	server2 := createTestServer(addrs[1], tempDirs[1], []string{addrs[0]})
	server3 := createTestServer(addrs[2], tempDirs[2], []string{addrs[0], addrs[1]})

	// Start servers
	servers := []*FileServer{server1, server2, server3}
	for i, server := range servers {
		go func(s *FileServer, idx int) {
			err := s.Start()
			if err != nil {
				t.Logf("Server %d error: %v", idx+1, err)
			}
		}(server, i)
		time.Sleep(100 * time.Millisecond)
	}

	// Give time for all connections to establish
	time.Sleep(1 * time.Second)

	// Store multiple files
	testFiles := map[string][]byte{
		"file1.txt": []byte("Content of file 1"),
		"file2.txt": []byte("Content of file 2"),
		"file3.txt": []byte("Content of file 3"),
	}

	// Store files on different servers
	i := 0
	for key, data := range testFiles {
		server := servers[i%len(servers)]
		err := server.Store(key, bytes.NewReader(data))
		assert.Nil(t, err)
		i++
	}

	// Give time for replication
	time.Sleep(500 * time.Millisecond)

	// Verify all files can be retrieved from all servers
	for key, expectedData := range testFiles {
		for j, server := range servers {
			reader, err := server.Get(key)
			assert.Nil(t, err, "Server %d should be able to get %s", j+1, key)

			data, err := io.ReadAll(reader)
			assert.Nil(t, err)
			assert.Equal(t, expectedData, data, "Data mismatch for %s on server %d", key, j+1)
		}
	}

	// Clean up
	for _, server := range servers {
		server.Stop()
	}
	time.Sleep(100 * time.Millisecond)
}

func TestFileServerNetworkFailure(t *testing.T) {
	// Create temporary directories
	tempDir1 := "/tmp/fs_fail_1"
	tempDir2 := "/tmp/fs_fail_2"
	defer func() {
		os.RemoveAll(tempDir1)
		os.RemoveAll(tempDir2)
	}()

	// Get free ports
	listener1, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	addr1 := listener1.Addr().String()
	listener1.Close()

	listener2, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	addr2 := listener2.Addr().String()
	listener2.Close()

	// Create servers
	server1 := createTestServer(addr1, tempDir1, []string{})
	server2 := createTestServer(addr2, tempDir2, []string{addr1})

	// Start only server1
	go func() {
		server1.Start()
	}()
	time.Sleep(100 * time.Millisecond)

	// Store file on server1
	testData := []byte("Test data for network failure")
	testKey := "network_test.txt"
	
	err = server1.Store(testKey, bytes.NewReader(testData))
	assert.Nil(t, err)

	// Try to get file from server2 (should fail gracefully)
	_, err = server2.Get(testKey)
	assert.NotNil(t, err, "Should fail when no network connection exists")

	// Clean up
	server1.Stop()
	server2.Stop()
}

func createTestServer(listenAddr, storageRoot string, bootstrapNodes []string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       storageRoot,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
	}

	server := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = server.OnPeer

	return server
}

func BenchmarkFileStorage(b *testing.B) {
	tempDir := "/tmp/fs_bench"
	defer os.RemoveAll(tempDir)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		b.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()

	server := createTestServer(addr, tempDir, []string{})
	go server.Start()
	time.Sleep(100 * time.Millisecond)
	defer server.Stop()

	testData := bytes.Repeat([]byte("benchmark data "), 1000) // ~15KB

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_file_%d.txt", i)
			err := server.Store(key, bytes.NewReader(testData))
			if err != nil {
				b.Error(err)
			}
			i++
		}
	})
}

func BenchmarkFileRetrieval(b *testing.B) {
	tempDir := "/tmp/fs_bench_get"
	defer os.RemoveAll(tempDir)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		b.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()

	server := createTestServer(addr, tempDir, []string{})
	go server.Start()
	time.Sleep(100 * time.Millisecond)
	defer server.Stop()

	// Pre-populate with test files
	testData := bytes.Repeat([]byte("benchmark data "), 1000)
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		key := fmt.Sprintf("bench_get_file_%d.txt", i)
		server.Store(key, bytes.NewReader(testData))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_get_file_%d.txt", i%numFiles)
			reader, err := server.Get(key)
			if err != nil {
				b.Error(err)
				continue
			}
			_, err = io.ReadAll(reader)
			if err != nil {
				b.Error(err)
			}
			i++
		}
	})
}
