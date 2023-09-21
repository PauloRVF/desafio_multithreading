package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/PauloRVF/desafio_multithreading/zzz_desafio_mutithreading/dto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("You must inform a CEP. Example: 02310000")
		return
	}
	cep := os.Args[1]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chViaCep := make(chan dto.ViaCepResponse)
	chCdn := make(chan dto.CdnResponse)

	go doRequestViaCEP(ctx, cep, chViaCep)
	go doRequestCDN(ctx, cep, chCdn)

	select {
	case viaCepResponse := <-chViaCep:
		fmt.Println("ViaCepResponse: ", viaCepResponse)
	case cdnResponse := <-chCdn:
		fmt.Println("CdnResponse: ", cdnResponse)
	case <-time.After(time.Second * 1):
		fmt.Println("Timeout")
	}
}

func doFormatCEP(cep string) string {
	return fmt.Sprintf("%s-%s", cep[:5], cep[5:])
}

func doRequestViaCEP(ctx context.Context, cep string, channel chan<- dto.ViaCepResponse) {
	response, err := doRequest(ctx, "https://viacep.com.br/ws/"+cep+"/json/")
	if err != nil {
		fmt.Printf("ViaCEPError: %s\n", err.Error())
		return
	}

	var viaCepResponse dto.ViaCepResponse
	err = json.Unmarshal(response, &viaCepResponse)
	if err != nil {
		fmt.Printf("ViaCEPResponseParseError: %s\n", err.Error())
		return
	}

	channel <- viaCepResponse
}

func doRequestCDN(ctx context.Context, cep string, channel chan<- dto.CdnResponse) {
	cep = doFormatCEP(cep)
	response, err := doRequest(ctx, "https://cdn.apicep.com/file/apicep/"+cep+".json")
	if err != nil {
		fmt.Printf("CDNError: %s\n", err.Error())
		return
	}

	var cdnResponse dto.CdnResponse
	err = json.Unmarshal(response, &cdnResponse)
	if err != nil {
		fmt.Printf("CDNResponseParseError: %s\n", err.Error())
		return
	}

	channel <- cdnResponse
}

func doRequest(ctx context.Context, URI string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", URI, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("APIError %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
