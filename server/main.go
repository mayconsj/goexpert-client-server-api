package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRateResponse struct {
	USDBRL struct {
		Code       string `json:"code"`
		CodeIn     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type ExchangeRate struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

const (
	dbPath = "./exchange_rates.db"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %s\n", err)
		panic(err)
	}

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS exchange_rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			Code TEXT,
			CodeIn TEXT,
			Name TEXT,
			High TEXT,
			Low TEXT,
			VarBid TEXT,
			PctChange TEXT,
			Bid TEXT,
			Ask TEXT,
			Timestamp TEXT,
			CreateDate TEXT
		);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("Error creating table: %s\n", err)
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	exchangeRate, err := getExchangeRate()
	if err != nil {
		log.Println("Error getting exchange rate:", err)
		http.Error(w, fmt.Sprintf("Error getting exchange rate: %s", err), http.StatusInternalServerError)
		return
	}

	err = insertExchangeRate((*ExchangeRate)(&exchangeRate.USDBRL))
	if err != nil {
		log.Println("Error persisting exchange rate:", err)
		http.Error(w, fmt.Sprintf("Error persisting exchange rate: %s", err), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(exchangeRate.USDBRL.Bid)
	if err != nil {
		log.Println("Error serializing JSON response:", err)
		http.Error(w, fmt.Sprintf("Error serializing JSON response: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonResponse)
}

func getExchangeRate() (*ExchangeRateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if error != nil {
		return nil, error
	}

	resp, error := http.DefaultClient.Do(req)
	if error != nil {
		return nil, error
	}

	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	var exchangeRate ExchangeRateResponse
	error = json.Unmarshal(body, &exchangeRate)
	if error != nil {
		return nil, error
	}

	return &exchangeRate, nil
}

func insertExchangeRate(rate *ExchangeRate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	insertSQL := `
	INSERT INTO exchange_rates (Code, CodeIn, Name, High, Low, VarBid, PctChange, Bid, Ask, Timestamp, CreateDate)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err := db.ExecContext(ctx, insertSQL, rate.Code, rate.CodeIn, rate.Name, rate.High, rate.Low, rate.VarBid, rate.PctChange, rate.Bid, rate.Ask, rate.Timestamp, rate.CreateDate)
	return err
}

func main() {
	defer db.Close()
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}
