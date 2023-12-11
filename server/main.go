package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotation struct {
	ForexRate ForexRate
}

type ForexRate struct {
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
}

func main() {
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, req *http.Request) {
		ctxHttp, cancelHttp := context.WithTimeout(context.Background(), 200*time.Millisecond)

		defer cancelHttp()

		cotation, err := fetchCotation(ctxHttp, "USD-BRL")

		if err != nil {
			log.Printf("Erro ao solicitar a cotação: %v\n", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		bidResponse := map[string]string{
			"bid": cotation.ForexRate.Bid,
		}

		err = json.NewEncoder(w).Encode(bidResponse)
		if err != nil {
			log.Printf("Erro ao codificar a resposta: %v\n", err)
			http.Error(w, `{"error": "Erro interno"}`, http.StatusInternalServerError)
		}
		ctxDb, cancelDb := context.WithTimeout(context.Background(), 10*time.Millisecond)

		defer cancelDb()

		storeCotation(ctxDb, &cotation.ForexRate)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fetchCotation(ctx context.Context, forexPair string) (*Cotation, error) {
	client := &http.Client{}
	request, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/"+forexPair, nil)

	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var responseMap map[string]ForexRate

	if err := json.Unmarshal(body, &responseMap); err != nil {
		return nil, err
	}

	jsonKey := strings.Replace(forexPair, "-", "", -1)

	return &Cotation{ForexRate: responseMap[jsonKey]}, nil
}

func storeCotation(ctx context.Context, forexRate *ForexRate) error {
	db, err := sql.Open("sqlite3", "./database.db")

	if err != nil {
		log.Printf("Erro ao abrir o banco de dados: %v\n", err)
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(
		`INSERT INTO cotations(Code, Codein, Name, High, Low, VarBid, PctChange, Bid, Ask, Timestamp, CreateDate)
			 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Erro ao preparar a declaração no banco: %v\n", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, forexRate.Code, forexRate.Codein, forexRate.Name,
		forexRate.High, forexRate.Low, forexRate.VarBid, forexRate.PctChange,
		forexRate.Bid, forexRate.Ask, forexRate.Timestamp, forexRate.CreateDate)
	if err != nil {
		log.Printf("Erro ao tentar salvar: %v\n", err)
		return err
	}

	return nil
}
