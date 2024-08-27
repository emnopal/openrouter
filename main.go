package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/joho/godotenv"
)

const (
	URL   = "https://openrouter.ai/api/v1/chat/completions"
	MODEL = "anthropic/claude-3-haiku"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type APIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func GetAPIKey() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	API_KEY := os.Getenv("API_KEY")

	if API_KEY == "" {
		log.Fatal("API_KEY is not set")
	}
	return API_KEY
}

func Request(requestBody APIRequest) ([]byte, error) {
	API_KEY := GetAPIKey()

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

func FetchLLM(prompt string) (Message, error) {
	messages := GenerateLLMPrompt(prompt)
	requestBody := APIRequest{
		Model:    MODEL,
		Messages: messages,
	}

	body, err := Request(requestBody)
	if err != nil {
		return Message{}, err
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return Message{}, fmt.Errorf("error unmarshaling response: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return Message{}, fmt.Errorf("no choices in API response")
	}

	return apiResponse.Choices[0].Message, nil
}

func GenerateLLMPrompt(prompt string) []Message {
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant"},
	}
	messages = append(messages, Message{Role: "user", Content: prompt})
	return messages
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		println(err.Error())
		return
	}
	result, err := FetchLLM(text)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("%s:\n\n %s\n", result.Role, markdown.Render(result.Content, 120, 0))
}
