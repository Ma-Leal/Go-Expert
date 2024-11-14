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

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
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

func main() {
	fmt.Println("Starting Server ...")

	db, err := sql.Open("sqlite3", "./desafio.db")
	if err != nil {
		fmt.Printf("Erro ao conectar com o SQLite: %v\n", err)
		return
	}
	defer db.Close()
	createTable(db)

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		HandlerServer(w, r, db)
	})

	http.ListenAndServe(":8080", nil)
}

func HandlerServer(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		fmt.Printf("Erro ao criar requisição: %v\n", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro: request ultrapassado o limite de 200ms -  %v", err)

		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler corpo da resposta: %v\n", err)
		return
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Printf("Erro no parse do JSON: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.USDBRL)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer dbCancel()

	if err := writeInDB(dbCtx, cotacao, db); err != nil {
		fmt.Printf("Erro ao salvar cotação no banco: %v\n", err)
		return
	}

}

func writeInDB(ctx context.Context, cotacao Cotacao, db *sql.DB) error {
	query := `INSERT INTO cotacoes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		cotacao.USDBRL.Code,
		cotacao.USDBRL.Codein,
		cotacao.USDBRL.Name,
		cotacao.USDBRL.High,
		cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid,
		cotacao.USDBRL.PctChange,
		cotacao.USDBRL.Bid,
		cotacao.USDBRL.Ask,
		cotacao.USDBRL.Timestamp,
		time.Now(),
	)

	return err
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		varBid TEXT,
		pctChange TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v\n", err)
	}
}
