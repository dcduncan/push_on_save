package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

type Payload struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type Response struct {
	Response string `json:"response"`
}

func main() {
	for {
		time.Sleep(1 * time.Second)
		hasChanges, err := hasChanges()
		if err != nil {
			panic(err)
		}

		if !hasChanges {
			fmt.Printf("No changes detected\n")
			continue
		}

		err = addChanges()
		if err != nil {
			panic(err)
		}

		diff, err := getDiff()
		if err != nil {
			panic(err)
		}

		message, err := generateCommitMessage(diff)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Committing with message %s\n", message)

		err = commit(message)
		if err != nil {
			panic(err)
		}

		err = push()
		if err != nil {
			panic(err)
		}
	}
}

func push() error {
	_, err := executeCommand("git", "push", "-u", "origin", "main")
	if err != nil {
		return fmt.Errorf("error pushing: %w", err)
	}

	return nil
}

func commit(message string) error {
	_, err := executeCommand("git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("error committing: %w", err)
	}
	return nil
}

func addChanges() error {
	_, err := executeCommand("git", "add", ".")
	return err
}

func hasChanges() (bool, error) {
	output, err := executeCommand("git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("error checking for changes: %w", err)
	}

	return len(output) > 0, nil
}

func executeCommand(command string, args ...string) (string, error) {

	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command (%s) failed: %w, stderr: %s", command, err, stderr.String())
	}

	return out.String(), nil

}

func getDiff() (string, error) {
	return executeCommand("git", "diff", "--staged")
}

func generateCommitMessage(diff string) (string, error) {
	url := "http://localhost:11435/api/generate"
	method := "POST"
	payload := Payload{
		Model:  "llama3",
		System: "You will receive a git diff. You will generate a good commit message from this diff. It must stay within 60 characters. Only output the commit message and nothing more.",
		Prompt: diff,
		Stream: false,
	}

	jsonData, err := json.Marshal(payload)

	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("the HTTP request failed with error %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	return response.Response, nil
}
