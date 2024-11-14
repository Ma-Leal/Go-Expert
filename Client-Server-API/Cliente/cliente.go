package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	fmt.Println("Starting Client ...")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Printf("Erro ao criar requisição: %v\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Erro: ultrapassado o limite de 300ms : %v\n", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Erro no read Body: %v\n", err)
		return
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		fmt.Printf("Erro no parse: %v\n", err)
		return
	}
	SaveInFile(cotacao.Bid)

	fmt.Printf("Cotação do dólar: $%v\n", cotacao.Bid)

}

func SaveInFile(valor string) {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Printf("Erro ao criar arquivo cotocao.txt: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "Dólar: {%v}\n", valor)
}
