package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Create .ssh directory
	sshDir := "/root/.ssh"
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		log.Fatalf("Failed to create .ssh directory: %v", err)
	}

	// Write environment file
	envPath := filepath.Join(sshDir, "environment")
	envFile, err := os.Create(envPath)
	if err != nil {
		log.Fatalf("Failed to create environment file: %v", err)
	}
	defer envFile.Close()

	// Environment variables to preserve
	envVars := []string{
		"ASSIGNMENT_TITLE",
		"AUTHENTICATION_TOKEN",
		"BASIC_AUTH",
		"SUBMISSION_URL",
		"SUBMIT_RESULTS_URL",
		"DEVEL",
	}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			fmt.Fprintf(envFile, "%s=%s\n", envVar, value)
		}
	}

	// Generate SSH host keys
	log.Println("Generating SSH host keys...")
	keygenCmd := exec.Command("/usr/bin/ssh-keygen", "-A")
	if err := keygenCmd.Run(); err != nil {
		log.Printf("Warning: Failed to generate SSH keys: %v", err)
	}

	// Start SSH daemon
	log.Println("Starting SSH daemon...")
	sshdCmd := exec.Command("/usr/sbin/sshd", "-D")
	sshdCmd.Stdout = os.Stdout
	sshdCmd.Stderr = os.Stderr

	if err := sshdCmd.Run(); err != nil {
		log.Fatalf("SSH daemon exited with error: %v", err)
	}
}
