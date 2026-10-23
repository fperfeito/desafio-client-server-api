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

var db *sql.DB

type APIResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalln("erro ao conectar ao banco de dados: ", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			cotacao TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalln("erro ao criar tabela cotacoes: ", err)
	}

	fmt.Println("iniciando server... aguardando requisições porta 8080 ...")
	http.HandleFunc("/cotacao", cotacao)
	http.ListenAndServe(":8080", nil)
}

func cotacao(w http.ResponseWriter, r *http.Request) {

	Bid, err := getCotacao()
	if err != nil {
		log.Println("error: não foi possível obter a cotação", err)
		http.Error(w, "erro ao obter cotação: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := salvarCotacaoNoBanco(CotacaoResponse{Bid: Bid}); err != nil {
		log.Println("erro/timeout ao persistir no banco: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CotacaoResponse{Bid: Bid})
}

func getCotacao() (string, error) {
	var dadosAPICotacao APIResponse

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &dadosAPICotacao)
	if err != nil {
		return "", err
	}

	return dadosAPICotacao.USDBRL.Bid, nil
}

func salvarCotacaoNoBanco(cotacao CotacaoResponse) (bool, error) {

	ctxBD, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctxBD, `
		INSERT INTO cotacoes (data, cotacao)
		VALUES (datetime('now'), ?)
	`,
		cotacao.Bid,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}
