package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const URL_SERVER = "http://localhost:8080/quote"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if isRequestVerified(ctx) {
		fmt.Println("Server is not available.")
	}

	bid := requestOut(URL_SERVER, ctx)
	createFile(bid)
}

func isRequestVerified(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func requestOut(url string, ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error decoding the JSON: %s\n", err)
	}
	bid := result["bid"].(string)
	return bid
}

func createFile(line string) {
	file, err := os.Create("quote.txt")
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err)
		return
	}
	size, err := file.Write([]byte("Dolar: " + line))
	if err != nil {
		panic(err)
	}
	fmt.Printf("File created successfully! size: %d bytes\n", size)
	file.Close()
	if err != nil {
		fmt.Printf("Error saving file: %s\n", err)
		return
	}
	fmt.Printf("Dollar quotation saved in quote.txt: Dólar: %s\n", line)
}
