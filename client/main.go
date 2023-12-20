package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Bid string `json:"bid"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	filename := "./cotacao.txt"
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, error := io.ReadAll(res.Body)
	if error != nil {
		panic(error)
	}

	error = writeToFile(filename, fmt.Sprintf("Dollar: %s", body))
	if error != nil {
		panic(error)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}

func writeToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Println(string(content))

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
