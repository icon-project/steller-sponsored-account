package main

import (
	"context"
	"encoding/json"

	"github.com/icon-project/steller-sponsored-account/soroban"
)

type request struct {
	Address string `json:"address"`
	Data    string `json:"data"`
}

type response struct {
	Hash string `json:"hash"`
}

func handleRequest(ctx context.Context, event json.RawMessage) (*response, error) {
	var req request
	if err := json.Unmarshal(event, &req); err != nil {
		return nil, err
	}
	data := &soroban.ExecuteSponsoredRequest{
		NetworkPassphrase: NetworkPassphrase,
		Address:           req.Address,
		Data:              req.Data,
		Key:               key,
	}
	res, err := sorobanClient.BeginSponsor(ctx, data)
	if err != nil {
		return nil, err
	}
	return &response{Hash: res.Hash}, nil
}
