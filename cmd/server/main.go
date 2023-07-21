package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	databaseName   = "quotes.db"
	createTableSQL = `
	CREATE TABLE IF NOT EXISTS quotes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
)

const URL_API_DOLAR = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type Quote struct {
	Bid string `json:"bid"`
}

func main() {
	http.HandleFunc("/quote", searchQuoteHandler)

	fmt.Println("Server HTTP started in http://localhost:8080/quote")
	http.ListenAndServe(":8080", nil)
}

func isRequestVerified(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func save(ctx context.Context, db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	if isRequestVerified(ctx) {
		return fmt.Errorf("Database is not responding.")
	}
	_, err := db.ExecContext(ctx, "INSERT INTO quotes (bid) VALUES (?)", bid)
	if err != nil {
		return err
	}

	return nil
}

func requestQuoteOut() (*Quote, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if isRequestVerified(ctx) {
		return nil, fmt.Errorf("API server not responding.")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", URL_API_DOLAR, nil)

	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	var data map[string]Quote
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	quote, ok := data["USDBRL"]
	if !ok {
		return nil, fmt.Errorf("quote not found in the response.")
	}

	return &quote, nil
}

func searchQuoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	quote, err := requestQuoteOut()
	if err != nil {
		http.Error(w, "Error getting quote.", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		http.Error(w, "Error to connecting in database.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec(createTableSQL)
	if err != nil {
		http.Error(w, "error creating tables", http.StatusInternalServerError)
		return
	}

	err = save(ctx, db, quote.Bid)
	if err != nil {
		http.Error(w, "Error saving quotation to database", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(quote)
}
