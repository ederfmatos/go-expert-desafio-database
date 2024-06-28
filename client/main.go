package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Exchange struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	handleError(err)
	response, err := http.DefaultClient.Do(request)
	handleError(err)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("status code error: %d %s", response.StatusCode, response.Status)
	}
	var exchange Exchange
	body, err := io.ReadAll(response.Body)
	handleError(err)
	err = json.Unmarshal(body, &exchange)
	handleError(err)
	file, err := os.Create("cotacao.txt")
	handleError(err)
	defer file.Close()
	_, err = file.Write([]byte("Dólar: " + exchange.Bid))
	handleError(err)
	log.Println("Exchange: ", exchange.Bid)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
