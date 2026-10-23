package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func gravaCotacao(cotacao string) (bool, error) {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return false, err
	}
	defer f.Close()

	tamanho, err := f.WriteString("Dólar: " + cotacao)
	if err != nil {
		return false, err
	}
	fmt.Printf("Arquivo criado com sucesso! Tamanho: %d bytes\n", tamanho)

	return true, nil
}

func getCotacao() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
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


	var cotacaoResp CotacaoResponse
	err = json.Unmarshal(body, &cotacaoResp)
	if err != nil {
		return "", err
	}

	return cotacaoResp.Bid, nil
}

func main() {

	cotacao, err := getCotacao()
	if err != nil {
		log.Fatal("erro ao obter cotação: ", err.Error())
	}

	_, err = gravaCotacao(cotacao)
	if err != nil {
		log.Fatal("erro ao gravar cotação: ", err.Error())
	}

	log.Println("Cotação obtida com sucesso! Valor da cotação: ", cotacao)

}
