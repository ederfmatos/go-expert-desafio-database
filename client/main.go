package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
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
	log.Println("Exchange: ", exchange.Bid)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
