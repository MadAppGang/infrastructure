package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runCommandWithOutput(name string, args ...string) (string, error) {
	fmt.Println("▶️ runCommandWithOutput:", name, args)
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating stdout pipe:", err)
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating stderr pipe:", err)
		return "", err
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		return "", err
	}

	// Create channels for output capture
	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)

	// Start goroutine to read from stdout and capture it
	go func() {
		output := streamOutputAndCapture(stdout, "STDOUT")
		stdoutChan <- output
	}()

	// Start goroutine to read from stderr and capture error text
	go func() {
		output := streamOutputAndCapture(stderr, "STDERR")
		stderrChan <- output
	}()

	// Wait for both outputs
	stdoutBuffer := <-stdoutChan
	stderrBuffer := <-stderrChan

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Command finished with error:", err)
		return stderrBuffer, err
	}
	return stdoutBuffer, nil
}

func streamOutput(r io.Reader, prefix string, doneChan chan bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("%s: %s\n", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("%s: Error reading output: %s\n", prefix, err)
	}
	doneChan <- true
}

func streamOutputAndCapture(r io.Reader, prefix string) string {
	var buffer strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("%s: %s\n", prefix, line)
		buffer.WriteString(line + "\n")
	}
	return buffer.String()
}

func CopyFolder(source, destination string) error {
	err := os.MkdirAll(destination, 0o755)
	if err != nil {
		return fmt.Errorf("error creating destination folder: %v", err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("error reading source folder: %v", err)
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(destination, entry.Name())

		if entry.IsDir() {
			err = CopyFolder(sourcePath, destPath)
			if err != nil {
				return fmt.Errorf("error copying subfolder: %v", err)
			}
		} else {
			err = CopyFile(sourcePath, destPath)
			if err != nil {
				return fmt.Errorf("error copying file: %v", err)
			}
		}
	}

	return nil
}

func CopyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file contents: %v", err)
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("error getting source file info: %v", err)
	}

	err = os.Chmod(destination, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error setting destination file permissions: %v", err)
	}

	return nil
}
