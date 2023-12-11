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

type Cotation struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cotation, err := fetchDolarCotation(ctx)

	if err != nil {
		log.Printf("Erro ao solicitar a cotação: %v\n", err)
		return
	}

	err = storeCotation(&cotation)
	if err != nil {
		log.Printf("Erro ao salvar a cotação: %v\n", err)
		return
	}

	log.Printf("Cotação do dolar: %s\n", cotation.Bid)
}

func fetchDolarCotation(ctx context.Context) (Cotation, error) {
	client := &http.Client{}

	var cotation Cotation

	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		return cotation, err
	}

	response, err := client.Do(request)

	if err != nil {
		return cotation, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return cotation, fmt.Errorf("Erro na resposta: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return cotation, err
	}

	err = json.Unmarshal(body, &cotation)

	if err != nil {
		return cotation, err
	}

	return cotation, nil
}

func storeCotation(cotation *Cotation) error {
	log.Printf("Salvando cotação: %s\n", cotation.Bid)

	file, err := os.Create("cotacao_atual.txt")

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dolár: %s\n", cotation.Bid))

	if err != nil {
		return err
	}

	return nil
}
