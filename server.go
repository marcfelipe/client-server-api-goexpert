package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	//"gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

/*
Com driver tradicional sqlite do gorm, apresentando erro CGO_ENABLED.
Na documentação do gorm há um exemplo comentado explicando que este driver não utiliza o CGO
https://gorm.io/docs/connecting_to_the_database.html
https://github.com/glebarez/sqlite

com os timeouts informados, programa não funciona. Porém com 2s para request e 1s para salvar no bd são suficientes
*/
const database string = "client-server-api.db"

type ApiResponse struct {
	Cotation Cotation `json:"USDBRL"`
}

type Cotation struct {
	ID         int    `gorm:"primarykey" json:"-"`
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

func OpenDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("cotacao_goexpert.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Cotation{})
	return db, nil
}

func main() {
	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	cotacaoAtual, err := BuscaoCotacao()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//InserirCotacao(cotacaoAtual)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacaoAtual)
}

func BuscaoCotacao() (*ApiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}
	defer cancel()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}
	var data ApiResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	InserirCotacao(data)
	return &data, nil
}

func InserirCotacao(cotacao ApiResponse) {
	db, err := OpenDb()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	if err := db.WithContext(ctx).Create(&Cotation{
		Code:       cotacao.Cotation.Code,
		Codein:     cotacao.Cotation.Codein,
		Name:       cotacao.Cotation.Name,
		High:       cotacao.Cotation.High,
		Low:        cotacao.Cotation.Low,
		VarBid:     cotacao.Cotation.VarBid,
		PctChange:  cotacao.Cotation.PctChange,
		Bid:        cotacao.Cotation.Bid,
		Ask:        cotacao.Cotation.Ask,
		Timestamp:  cotacao.Cotation.Timestamp,
		CreateDate: cotacao.Cotation.CreateDate,
	}).Error; err != nil {
		panic(err.Error())
	}
}
