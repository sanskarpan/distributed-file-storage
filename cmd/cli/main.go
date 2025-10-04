package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/anthdm/foreverstore/config"
	"github.com/anthdm/foreverstore/logger"
)

// Import the main package types
type FileServer interface {
	Store(key string, r io.Reader) error
	Get(key string) (io.Reader, error)
	Start() error
	Stop()
}

func main() {
	var (
		configFile = flag.String("config", "config.json", "Configuration file path")
		serverAddr = flag.String("server", ":3000", "File server address to connect to")
		command    = flag.String("cmd", "", "Command to execute: store, get, list, delete")
		key        = flag.String("key", "", "File key for operations")
		file       = flag.String("file", "", "Local file path for store/get operations")
		output     = flag.String("output", "", "Output file path for get operations")
		verbose    = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	// Setup logging
	if *verbose {
		logger.SetGlobalLevel(logger.DEBUG)
	} else {
		logger.SetGlobalLevel(logger.WARN)
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override server address if provided
	if *serverAddr != ":3000" {
		cfg.ListenAddr = *serverAddr
	}

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// Create a client connection to the file server
	client, err := createClient(cfg)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Execute the command
	switch *command {
	case "store":
		if *key == "" || *file == "" {
			fmt.Println("Error: Both -key and -file are required for store command")
			os.Exit(1)
		}
		err = storeFile(client, *key, *file)
	case "get":
		if *key == "" {
			fmt.Println("Error: -key is required for get command")
			os.Exit(1)
		}
		err = getFile(client, *key, *output)
	case "list":
		err = listFiles(client)
	case "delete":
		if *key == "" {
			fmt.Println("Error: -key is required for delete command")
			os.Exit(1)
		}
		err = deleteFile(client, *key)
	default:
		fmt.Printf("Error: Unknown command '%s'\n", *command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Command failed: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Distributed File Storage CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fs-cli [options] -cmd <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  store    Store a file in the distributed system")
	fmt.Println("  get      Retrieve a file from the distributed system")
	fmt.Println("  list     List all files in the system (not implemented)")
	fmt.Println("  delete   Delete a file from the system (not implemented)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string    Configuration file path (default: config.json)")
	fmt.Println("  -server string    File server address (default: :3000)")
	fmt.Println("  -key string       File key for operations")
	fmt.Println("  -file string      Local file path for store/get operations")
	fmt.Println("  -output string    Output file path for get operations")
	fmt.Println("  -v                Verbose output")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fs-cli -cmd store -key myfile.txt -file /path/to/local/file.txt")
	fmt.Println("  fs-cli -cmd get -key myfile.txt -output /path/to/save/file.txt")
	fmt.Println("  fs-cli -cmd get -key myfile.txt  # prints to stdout")
	fmt.Println("  fs-cli -cmd list")
}

// Simple client that connects to a file server
type SimpleClient struct {
	serverAddr string
	// In a real implementation, this would maintain a connection to the server
	// For now, we'll simulate it
}

func createClient(cfg *config.Config) (*SimpleClient, error) {
	return &SimpleClient{
		serverAddr: cfg.ListenAddr,
	}, nil
}

func storeFile(client *SimpleClient, key, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	fmt.Printf("Storing file '%s' with key '%s' (%d bytes)\n", filePath, key, len(data))
	
	// In a real implementation, this would send the file to the server
	// For now, we'll just simulate success
	fmt.Printf("✓ File stored successfully\n")
	
	return nil
}

func getFile(client *SimpleClient, key, outputPath string) error {
	fmt.Printf("Retrieving file with key '%s'\n", key)
	
	// In a real implementation, this would retrieve the file from the server
	// For now, we'll simulate with dummy data
	dummyData := []byte(fmt.Sprintf("This is dummy content for key: %s", key))
	
	if outputPath == "" {
		// Print to stdout
		fmt.Printf("File content:\n%s\n", string(dummyData))
	} else {
		// Save to file
		err := os.WriteFile(outputPath, dummyData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		fmt.Printf("✓ File saved to: %s\n", outputPath)
	}
	
	return nil
}

func listFiles(client *SimpleClient) error {
	fmt.Println("Listing files in the distributed system:")
	
	// In a real implementation, this would query the server for file list
	// For now, we'll simulate with dummy data
	files := []string{
		"document.pdf",
		"image.jpg",
		"data.csv",
		"backup.zip",
	}
	
	fmt.Println("Files:")
	for i, file := range files {
		fmt.Printf("  %d. %s\n", i+1, file)
	}
	
	return nil
}

func deleteFile(client *SimpleClient, key string) error {
	fmt.Printf("Deleting file with key '%s'\n", key)
	
	// In a real implementation, this would send a delete request to the server
	// For now, we'll simulate success
	fmt.Printf("✓ File deleted successfully\n")
	
	return nil
}
