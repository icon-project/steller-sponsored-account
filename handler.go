package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/icon-project/steller-sponsored-account/soroban"
	"github.com/stellar/go/txnbuild"
)

type request struct {
	Address string `json:"address"`
	Data    string `json:"data"`
}

func handleRequest(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if event.RequestContext.HTTP.Method != "POST" {
		return events.LambdaFunctionURLResponse{
			StatusCode: 405,
			Body:       "Method Not Allowed",
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Allow":                       "POST",
			},
		}, nil
	}

	var req request
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		log.Printf("error unmarshalling request: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       "Invalid request",
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Allow":                       "POST",
			},
		}, nil
	}

	if req.Data != "" {
		xdr, err := txnbuild.TransactionFromXDR(req.Data)
		if err != nil {
			log.Printf("error unmarshalling xdr: %v", err)
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       "Invalid XDR",
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Allow":                       "POST",
				},
			}, nil
		}
		txXDR, ok := xdr.Transaction()
		if !ok {
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       "Invalid XDR",
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Allow":                       "POST",
				},
			}, nil
		}
		if txXDR.SourceAccount().AccountID != key.Address() {
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       "Source account does not match",
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Allow":                       "POST",
				},
			}, nil
		}
		res, err := sorobanClient.SubmitTransactionXDR(ctx, req.Data)
		if err != nil {
			log.Printf("error submitting transaction: %v", err)
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       "Source account does not match",
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Allow":                       "POST",
				},
			}, nil
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Body:       res.Hash,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Allow":                       "POST",
			},
		}, nil
	} else if req.Address != "" {
		data := &soroban.ExecuteSponsoredRequest{
			NetworkPassphrase: NetworkPassphrase,
			Address:           req.Address,
			Key:               key,
		}
		res, err := sorobanClient.BeginSponsor(ctx, data)
		if err != nil {
			log.Printf("error beginning sponsorship: %v", err)
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       "Error beginning sponsorship",
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Allow":                       "POST",
				},
			}, nil
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Body:       res,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Allow":                       "POST",
			},
		}, nil
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: 400,
		Body:       "Invalid request",
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Allow":                       "POST",
		},
	}, nil
}
