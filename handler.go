package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
)

type request struct {
	Address string `json:"address"`
	Data    string `json:"data"`
}

var headers = map[string]string{
	"Content-Type": "text/plain",
}

const (
	ErrorInvalidRequest   = "Invalid request"
	ErrorInvalidXDR       = "Invalid XDR"
	ErrorSourceMismatch   = "Source account does not match"
	ErrorSponsorship      = "Error beginning sponsorship"
	ErrorSubmitting       = "Error submitting transaction"
	ErrorMethodNotAllowed = "Method Not Allowed"
	ErrorNotFound         = "Not Found"
)

var allowedOps = []xdr.OperationType{
	xdr.OperationTypeCreateAccount,
	xdr.OperationTypeBeginSponsoringFutureReserves,
	xdr.OperationTypeEndSponsoringFutureReserves,
}

type route struct {
	Handler func(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse
}

var routes = map[string]map[string]route{
	"/": {
		http.MethodPost: {handlePost},
	},
}

func handleRequest(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if methodRoutes, ok := routes[req.RawPath]; ok {
		if route, ok := methodRoutes[req.RequestContext.HTTP.Method]; ok {
			return route.Handler(ctx, req), nil
		}
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: 404,
		Body:       ErrorNotFound,
		Headers:    headers,
	}, nil
}

func handlePost(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse {
	var body request
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		log.Printf("error unmarshalling request: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorInvalidRequest,
			Headers:    headers,
		}
	}
	xdrs, err := txnbuild.TransactionFromXDR(body.Data)
	if err != nil {
		log.Printf("error unmarshalling xdr: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorInvalidXDR,
			Headers:    headers,
		}
	}
	txXDR, ok := xdrs.Transaction()
	if !ok {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorInvalidXDR,
			Headers:    headers,
		}
	}
	if txXDR.SourceAccount().AccountID != key.Address() {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorSourceMismatch,
			Headers:    headers,
		}
	}
	for _, op := range txXDR.Operations() {
		xdrOp, err := op.BuildXDR()
		if err != nil {
			log.Printf("error building xdr: %v", err)
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       ErrorInvalidXDR,
				Headers:    headers,
			}
		}
		switch xdrOp.Body.Type {
		case xdr.OperationTypeCreateAccount:
			if xdrOp.Body.CreateAccountOp.StartingBalance != 0 {
				return events.LambdaFunctionURLResponse{
					StatusCode: 400,
					Body:       ErrorInvalidXDR,
					Headers:    headers,
				}
			}
		case xdr.OperationTypeBeginSponsoringFutureReserves:
			continue
		case xdr.OperationTypeEndSponsoringFutureReserves:
			continue
		default:
			return events.LambdaFunctionURLResponse{
				StatusCode: 400,
				Body:       ErrorInvalidXDR,
				Headers:    headers,
			}
		}
	}
	signedXDR, err := txXDR.Sign(NetworkPassphrase, key)
	if err != nil {
		log.Printf("error signing xdr: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	xdrTxBase64, err := signedXDR.Base64()
	if err != nil {
		log.Printf("error encoding xdr: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	res, err := sorobanClient.SubmitTransactionXDR(ctx, xdrTxBase64)
	if err != nil {
		log.Printf("error submitting transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSubmitting,
			Headers:    headers,
		}
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       res.Hash,
		Headers:    headers,
	}
}
