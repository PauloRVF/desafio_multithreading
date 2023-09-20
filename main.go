package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	chViaCep := make(chan interface{})
	chCdn := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go doRequestViaCEP(ctx, "02310000", chViaCep)
	go doRequestCDN(ctx, "02310-000", chCdn)

	select {
	case viaCepResponse := <-chViaCep:
		fmt.Println("ViaCepResponse: ", viaCepResponse)
	case cdnResponse := <-chCdn:
		fmt.Println("CdnResponse: ", cdnResponse)
	case <-time.After(time.Second * 10):
		fmt.Println("Timeout")
	}
}

func doRequestViaCEP(ctx context.Context, cep string, channel chan<- interface{}) {
	response, err := doRequest(ctx, "https://viacep.com.br/ws/"+cep+"/json/")
	if err != nil {
		fmt.Printf("ViaCEPError: %s\n", err.Error())
		return
	}

	channel <- response
}

func doRequestCDN(ctx context.Context, cep string, channel chan<- interface{}) {
	response, err := doRequest(ctx, "https://cdn.apicep.com/file/apicep/"+cep+".json")
	if err != nil {
		fmt.Printf("CDNError: %s\n", err.Error())
		return
	}

	channel <- response
}

func doRequest(ctx context.Context, URI string) (map[string]interface{}, error) {
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

	var mapResponse map[string]interface{}
	err = json.Unmarshal(body, &mapResponse)
	if err != nil {
		return nil, err
	}

	return mapResponse, nil
}
