package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aymanbagabas/go-pty"
)

func main() {
	fmt.Println("PTY Command Execution Examples:")
	fmt.Println("1. Execute simple command")
	fmt.Println("2. Execute command with arguments")
	fmt.Println("3. Execute interactive shell")
	fmt.Println("4. Execute multiple commands")
	
	var choice int
	fmt.Print("Enter your choice (1-4): ")
	fmt.Scanf("%d", &choice)
	
	switch choice {
	case 1:
		executeSimpleCommand()
	case 2:
		executeCommandWithArgs()
	case 3:
		executeInteractiveShell()
	case 4:
		executeMultipleCommands()
	default:
		fmt.Println("Invalid choice, running simple command...")
		executeSimpleCommand()
	}
}

// Example 1: Execute a simple command
func executeSimpleCommand() {
	fmt.Println("\n=== Executing Simple Command (dir) ===")
	
	pty, err := pty.New()
	if err != nil {
		log.Fatalf("failed to open pty: %s", err)
	}
	defer pty.Close()
	
	// Execute 'dir' command (Windows) or 'ls' (Unix)
	c := pty.Command("cmd", "/c", "dir")
	if err := c.Start(); err != nil {
		log.Fatalf("failed to start: %s", err)
	}
	
	// Copy output to stdout
	go io.Copy(os.Stdout, pty)
	
	if err := c.Wait(); err != nil {
		log.Printf("Command finished with error: %v", err)
	}
}

// Example 2: Execute command with arguments
func executeCommandWithArgs() {
	fmt.Println("\n=== Executing Command with Arguments ===")
	
	pty, err := pty.New()
	if err != nil {
		log.Fatalf("failed to open pty: %s", err)
	}
	defer pty.Close()
	
	// Execute 'ping' command with arguments
	c := pty.Command("ping", "-n", "3", "google.com")
	if err := c.Start(); err != nil {
		log.Fatalf("failed to start: %s", err)
	}
	
	go io.Copy(os.Stdout, pty)
	
	if err := c.Wait(); err != nil {
		log.Printf("Command finished with error: %v", err)
	}
}

// Example 3: Execute interactive shell
func executeInteractiveShell() {
	fmt.Println("\n=== Interactive Shell (type 'exit' to quit) ===")
	
	pty, err := pty.New()
	if err != nil {
		log.Fatalf("failed to open pty: %s", err)
	}
	defer pty.Close()
	
	// Start PowerShell or bash
	c := pty.Command("powershell.exe")
	if err := c.Start(); err != nil {
		log.Fatalf("failed to start: %s", err)
	}
	
	// Handle input from user
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := scanner.Text()
			if strings.TrimSpace(input) == "exit" {
				pty.Write([]byte("exit\r\n"))
				break
			}
			pty.Write([]byte(input + "\r\n"))
		}
	}()
	
	// Copy output to stdout
	go io.Copy(os.Stdout, pty)
	
	if err := c.Wait(); err != nil {
		log.Printf("Shell finished with error: %v", err)
	}
}

// Example 4: Execute multiple commands sequentially
func executeMultipleCommands() {
	fmt.Println("\n=== Executing Multiple Commands ===")
	
	commands := [][]string{
		{"cmd", "/c", "echo Hello from PTY"},
		{"cmd", "/c", "date /t"},
		{"cmd", "/c", "time /t"},
		{"cmd", "/c", "echo Current directory:"},
		{"cmd", "/c", "cd"},
	}
	
	for i, cmdArgs := range commands {
		fmt.Printf("\n--- Command %d: %s ---\n", i+1, strings.Join(cmdArgs, " "))
		
		pty, err := pty.New()
		if err != nil {
			log.Fatalf("failed to open pty: %s", err)
		}
		
		c := pty.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := c.Start(); err != nil {
			log.Fatalf("failed to start: %s", err)
		}
		
		go io.Copy(os.Stdout, pty)
		
		if err := c.Wait(); err != nil {
			log.Printf("Command %d finished with error: %v", i+1, err)
		}
		
		pty.Close()
	}
}

// Helper function to execute a single command and return output
func ExecuteCommandAndGetOutput(command string, args ...string) (string, error) {
	pty, err := pty.New()
	if err != nil {
		return "", fmt.Errorf("failed to open pty: %w", err)
	}
	defer pty.Close()
	
	c := pty.Command(command, args...)
	if err := c.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}
	
	// Read all output
	output, err := io.ReadAll(pty)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}
	
	if err := c.Wait(); err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}
	
	return string(output), nil
}