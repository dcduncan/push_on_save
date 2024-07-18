package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os/exec"
)

type Payload struct { 
  Model string `json:"model"`
  System string `json:"system"`
  Prompt string `json:"prompt"`
  Stream bool `json:"stream"`
}

type Response struct {
  Response string `json:"response"`
}


func main() {
  diff, err := getDiff()
  if err != nil {
    fmt.Printf("Error: %s\n", err)
    return
  }

  message, err := generateCommitMessage(diff)
  if err != nil {
    fmt.Printf("Error: %s\n", err)
    return
  }
  fmt.Printf(message)
}

func getDiff() (string, error) {
  cmd := exec.Command("git", "diff", "--staged")
  var out bytes.Buffer
  cmd.Stdout = &out
  var stderr bytes.Buffer
  cmd.Stderr = &stderr
  err := cmd.Run()
  if err != nil {
    return "", fmt.Errorf("git diff --staged failed: %w, stderr: %s", err, stderr.String())
  }

  return out.String(), nil
}

func generateCommitMessage(diff string) (string, error) {
  url := "http://localhost:11434/api/generate"
  method := "POST"
  payload := Payload {
    Model: "llama3",
    System: "You will receive a git diff. You will generate a good commit message from this diff. It must stay within 60 characters. Only output the commit message and nothing more.",
    Prompt: diff,
    Stream: false,
  }

  jsonData, err := json.Marshal(payload)

  if err != nil {
    return "", fmt.Errorf("Error marshalling JSON: %w", err)
  }

  fmt.Println(string(jsonData))
  req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
  if err != nil {
    return "", fmt.Errorf("Error creating request: %w", err)
  }

  req.Header.Set("Content-Type", "application/json")
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return "", fmt.Errorf("The HTTP request failed with error %w", err)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", fmt.Errorf("Error reading response: %w", err)
  }

  var response Response
  err = json.Unmarshal(body, &response)
  if err != nil {
    return "", fmt.Errorf("Error unmarshalling response: %w", err)
  }

  return response.Response, nil
}
