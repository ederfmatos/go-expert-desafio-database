package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DatabaseTimeout = 100 * time.Millisecond
const RequestTimeout = 2000 * time.Millisecond

func main() {
	server := http.NewServeMux()
	database := createDatabase()
	defer database.Close()
	server.HandleFunc("GET /cotacao", HandleExchange(database))
	log.Fatal(http.ListenAndServe(":8080", server))
}

func createDatabase() *sql.DB {
	database, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Erro ao abrir banco de dados: %v", err)
	}
	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS exchanges(
		    id INTEGER PRIMARY KEY AUTOINCREMENT, 
		    bid DECIMAL(6, 2) NOT NULL, 
		    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
	return database
}

func HandleExchange(database *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		exchange, err := GetExchange(request.Context())
		if err != nil {
			handleError(writer, err)
			return
		}

		databaseContext, cancel := context.WithTimeout(request.Context(), DatabaseTimeout)
		defer cancel()
		err = saveExchange(database, databaseContext, exchange)
		if err != nil {
			handleError(writer, err)
			return
		}
		err = json.NewEncoder(writer).Encode(exchange)
		if err != nil {
			handleError(writer, err)
			return
		}
	}
}

func handleError(writer http.ResponseWriter, err error) {
	log.Fatalf("An error occured: %v", err)
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func saveExchange(database *sql.DB, ctx context.Context, exchange *Exchange) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := database.ExecContext(ctx, "INSERT INTO exchanges (bid) VALUES (?)", exchange.Bid)
		return err
	}
}

type Response struct {
	USDBRL Exchange `json:"USDBRL"`
}

type Exchange struct {
	Bid string `json:"bid"`
}

func GetExchange(ctx context.Context) (*Exchange, error) {
	requestContext, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var exchange Response
	err = json.Unmarshal(body, &exchange)
	return &exchange.USDBRL, err
}
