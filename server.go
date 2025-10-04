package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/anthdm/foreverstore/errors"
	"github.com/anthdm/foreverstore/logger"
	"github.com/anthdm/foreverstore/p2p"
	"github.com/anthdm/foreverstore/retry"
)

type FileServerOpts struct {
	ID                string
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
	logger *logger.Logger
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}

	// Create a logger with the server's transport address as prefix
	serverLogger := logger.WithPrefix(fmt.Sprintf("SERVER[%s]", opts.Transport.Addr()))

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
		logger:         serverLogger,
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return errors.Wrap(err, errors.InternalError, "failed to encode broadcast message")
	}

	s.logger.Debug("Broadcasting message to %d peers", len(s.peers))
	
	var lastErr error
	successCount := 0
	
	for addr, peer := range s.peers {
		if err := peer.Send([]byte{p2p.IncomingMessage}); err != nil {
			s.logger.Warn("Failed to send message header to peer %s: %v", addr, err)
			lastErr = err
			continue
		}
		
		if err := peer.Send(buf.Bytes()); err != nil {
			s.logger.Warn("Failed to send message body to peer %s: %v", addr, err)
			lastErr = err
			continue
		}
		
		successCount++
	}

	if successCount == 0 && len(s.peers) > 0 {
		return errors.Wrap(lastErr, errors.NetworkError, "failed to broadcast to any peers")
	}
	
	if successCount < len(s.peers) {
		s.logger.Warn("Broadcast partially failed: %d/%d peers reached", successCount, len(s.peers))
	} else {
		s.logger.Debug("Broadcast successful to all %d peers", successCount)
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	// Check if file exists locally first
	if s.store.Has(s.ID, key) {
		s.logger.Info("Serving file (%s) from local disk", key)
		_, r, err := s.store.Read(s.ID, key)
		if err != nil {
			return nil, errors.Wrap(err, errors.StorageError, "failed to read local file")
		}
		return r, nil
	}

	s.logger.Info("File (%s) not found locally, fetching from network", key)

	// Use retry logic for network operations
	err := retry.DoSimple(func() error {
		return s.fetchFileFromNetwork(key)
	})
	
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "failed to fetch file from network")
	}

	// Read the file after successful network fetch
	_, r, err := s.store.Read(s.ID, key)
	if err != nil {
		return nil, errors.Wrap(err, errors.StorageError, "failed to read file after network fetch")
	}
	
	return r, nil
}

func (s *FileServer) fetchFileFromNetwork(key string) error {
	if len(s.peers) == 0 {
		return errors.NewNetworkError("no peers available for file retrieval")
	}

	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: hashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// Wait for responses with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return errors.NewTimeoutError("timeout waiting for file from network")
	case <-time.After(500 * time.Millisecond):
		// Continue with processing responses
	}

	var lastErr error
	for addr, peer := range s.peers {
		// First read the file size so we can limit the amount of bytes that we read
		// from the connection, so it will not keep hanging.
		var fileSize int64
		if err := binary.Read(peer, binary.LittleEndian, &fileSize); err != nil {
			s.logger.Warn("Failed to read file size from peer %s: %v", addr, err)
			lastErr = err
			continue
		}

		n, err := s.store.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			s.logger.Warn("Failed to write file from peer %s: %v", addr, err)
			lastErr = err
			peer.CloseStream()
			continue
		}

		s.logger.Info("Received (%d) bytes from peer %s", n, addr)
		peer.CloseStream()
		return nil // Success
	}

	if lastErr != nil {
		return lastErr
	}
	
	return errors.NewNetworkError("no peers provided the requested file")
}

func (s *FileServer) Store(key string, r io.Reader) error {
	s.logger.Info("Storing file: %s", key)
	
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	// Store file locally first
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return errors.Wrap(err, errors.StorageError, "failed to write file locally")
	}
	
	s.logger.Debug("File stored locally: %s (%d bytes)", key, size)

	// Only replicate if we have peers
	if len(s.peers) == 0 {
		s.logger.Warn("No peers available for replication")
		return nil
	}

	// Broadcast store message to peers
	msg := Message{
		Payload: MessageStoreFile{
			ID:   s.ID,
			Key:  hashKey(key),
			Size: size + 16, // Add encryption overhead
		},
	}

	if err := s.broadcast(&msg); err != nil {
		s.logger.Error("Failed to broadcast store message: %v", err)
		// Don't fail the entire operation if broadcast fails
	}

	// Small delay to ensure peers are ready
	time.Sleep(5 * time.Millisecond)

	// Replicate to all peers
	return s.replicateTopeers(key, fileBuffer)
}

func (s *FileServer) replicateTopeers(key string, fileBuffer *bytes.Buffer) error {
	if len(s.peers) == 0 {
		return nil
	}

	peers := make([]io.Writer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	
	mw := io.MultiWriter(peers...)
	
	// Send stream header
	if _, err := mw.Write([]byte{p2p.IncomingStream}); err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to send stream header")
	}
	
	// Encrypt and send file data
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return errors.Wrap(err, errors.EncryptionError, "failed to encrypt and send file data")
	}

	s.logger.Info("File replicated to %d peers (%d bytes)", len(s.peers), n)
	return nil
}

func (s *FileServer) Stop() {
	s.logger.Info("Stopping file server")
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	addr := p.RemoteAddr().String()
	s.peers[addr] = p

	s.logger.Info("Connected with peer: %s", addr)

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		s.logger.Info("File server stopped")
		s.Transport.Close()
	}()

	s.logger.Info("Starting message processing loop")
	
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				s.logger.Error("Failed to decode message from %s: %v", rpc.From, err)
				continue
			}
			
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				s.logger.Error("Failed to handle message from %s: %v", rpc.From, err)
			}

		case <-s.quitch:
			s.logger.Debug("Received quit signal")
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		s.logger.Debug("Handling store file message from %s", from)
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		s.logger.Debug("Handling get file message from %s", from)
		return s.handleMessageGetFile(from, v)
	default:
		s.logger.Warn("Unknown message type from %s", from)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.ID, msg.Key) {
		err := errors.NewFileNotFoundError(msg.Key)
		s.logger.Debug("File not found for peer %s: %s", from, msg.Key)
		return err
	}

	s.logger.Info("Serving file (%s) to peer %s", msg.Key, from)

	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return errors.Wrap(err, errors.StorageError, "failed to read file for serving")
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer func() {
			if err := rc.Close(); err != nil {
				s.logger.Warn("Failed to close file reader: %v", err)
			}
		}()
	}

	peer, ok := s.peers[from]
	if !ok {
		return errors.NewConnectionError(fmt.Sprintf("peer %s not found", from))
	}

	// First send the "incomingStream" byte to the peer and then we can send
	// the file size as an int64.
	if err := peer.Send([]byte{p2p.IncomingStream}); err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to send stream header")
	}
	
	if err := binary.Write(peer, binary.LittleEndian, fileSize); err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to send file size")
	}
	
	n, err := io.Copy(peer, r)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to send file data")
	}

	s.logger.Info("Sent file (%s) to peer %s: %d bytes", msg.Key, from, n)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return errors.NewConnectionError(fmt.Sprintf("peer %s not found", from))
	}

	s.logger.Debug("Receiving file from peer %s: %s (%d bytes)", from, msg.Key, msg.Size)

	n, err := s.store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return errors.Wrap(err, errors.StorageError, "failed to write file from peer")
	}

	s.logger.Info("Stored file from peer %s: %s (%d bytes)", from, msg.Key, n)

	peer.CloseStream()
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	if len(s.BootstrapNodes) == 0 {
		s.logger.Info("No bootstrap nodes configured")
		return nil
	}

	s.logger.Info("Bootstrapping network with %d nodes", len(s.BootstrapNodes))
	
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			s.logger.Info("Attempting to connect to bootstrap node: %s", addr)
			
			err := retry.DoSimple(func() error {
				return s.Transport.Dial(addr)
			})
			
			if err != nil {
				s.logger.Error("Failed to connect to bootstrap node %s: %v", addr, err)
			} else {
				s.logger.Info("Successfully connected to bootstrap node: %s", addr)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	s.logger.Info("Starting file server on %s", s.Transport.Addr())

	if err := s.Transport.ListenAndAccept(); err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to start transport listener")
	}

	if err := s.bootstrapNetwork(); err != nil {
		s.logger.Warn("Bootstrap network failed: %v", err)
	}

	s.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
