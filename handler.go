package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
)

type request struct {
	Data string `json:"data"`
}

var headers = map[string]string{
	"Content-Type": "text/plain",
}

const (
	ErrorInvalidRequest   = "Invalid request"
	ErrorInvalidXDR       = "Invalid XDR"
	ErrorSourceMismatch   = "Source account does not match"
	ErrorSponsorship      = "Error beginning sponsorship"
	ErrorRevoking         = "Error revoking sponsorship"
	ErrorNoSponsorship    = "No sponsorship found"
	ErrorSubmitting       = "Error submitting transaction"
	ErrorMethodNotAllowed = "Method Not Allowed"
	ErrorNotFound         = "Not Found"
)

type route struct {
	Handler func(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse
}

type listAccountResponse struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}

var routes = map[string]map[string]route{
	"/": {
		http.MethodPost: {handlePost},
	},
	"/revoke": {
		http.MethodGet:  {handleRevokeRequestAuto},
		http.MethodPost: {handleRevokeRequestFromClient},
	},
	"/list": {
		http.MethodGet: {handleListRequest},
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
	if len(txXDR.Operations()) != 3 {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorInvalidXDR,
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
	signedXDR, err := txXDR.Sign(networkPassphrase, key)
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

func handleRevokeRequestAuto(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse {
	accountInfo, err := sorobanClient.GetAccount(ctx, key.Address())
	if err != nil {
		log.Printf("error getting account info: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}

	limit := req.QueryStringParameters["limit"]
	if limit == "" {
		limit = "100"
	}
	order := req.QueryStringParameters["order"]
	if order == "" {
		order = "asc"
	}

	accounts, err := sorobanClient.GetCreateAccountOperation(ctx, key.Address(), limit, order)
	if err != nil {
		log.Printf("error getting account operations: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	if len(accounts) == 0 {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorNoSponsorship,
			Headers:    headers,
		}
	}

	revokeOps := []txnbuild.Operation{}
	for _, account := range accounts {
		for _, b := range account.Balance {
			if b.AssetType == "native" {
				if b.Balance == "0" {
					continue
				}
			}
		}
		revokeOps = append(revokeOps, &txnbuild.RevokeSponsorship{
			SourceAccount:   key.Address(),
			Account:         &account.AccountID,
			SponsorshipType: txnbuild.RevokeSponsorshipTypeAccount,
		})
	}
	sequence, err := strconv.ParseInt(accountInfo.Sequence, 10, 64)
	if err != nil {
		log.Printf("error parsing sequence: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	txParams := txnbuild.TransactionParams{
		SourceAccount:        &txnbuild.SimpleAccount{AccountID: accountInfo.AccountID, Sequence: sequence},
		Operations:           revokeOps,
		IncrementSequenceNum: true,
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	}
	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		log.Printf("error creating transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	signedTx, err := tx.Sign(networkPassphrase, key)
	if err != nil {
		log.Printf("error signing transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	xdrTxBase64, err := signedTx.Base64()
	if err != nil {
		log.Printf("error encoding transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	fmt.Println(xdrTxBase64)
	res, err := sorobanClient.SubmitTransactionXDR(ctx, xdrTxBase64)
	if err != nil {
		log.Printf("error submitting transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSubmitting,
			Headers:    headers,
		}
	}
	log.Printf("revoked sponsorship result: %v", res.ResultXdr)
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       res.Hash,
		Headers:    headers,
	}
}

func handleRevokeRequestFromClient(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse {
	accountInfo, err := sorobanClient.GetAccount(ctx, key.Address())
	if err != nil {
		log.Printf("error getting account info: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}

	var accounts []string

	if err := json.Unmarshal([]byte(req.Body), &accounts); err != nil {
		log.Printf("error unmarshalling request: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorInvalidRequest,
			Headers:    headers,
		}
	}

	if len(accounts) == 0 {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorNoSponsorship,
			Headers:    headers,
		}
	}
	revokeOps := []txnbuild.Operation{}
	for _, account := range accounts {
		revokeOps = append(revokeOps, &txnbuild.RevokeSponsorship{
			SourceAccount:   key.Address(),
			Account:         &account,
			SponsorshipType: txnbuild.RevokeSponsorshipTypeAccount,
		})
	}
	sequence, err := strconv.ParseInt(accountInfo.Sequence, 10, 64)
	if err != nil {
		log.Printf("error parsing sequence: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	txParams := txnbuild.TransactionParams{
		SourceAccount:        &txnbuild.SimpleAccount{AccountID: accountInfo.AccountID, Sequence: sequence},
		Operations:           revokeOps,
		IncrementSequenceNum: true,
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	}
	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		log.Printf("error creating transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	signedTx, err := tx.Sign(networkPassphrase, key)
	if err != nil {
		log.Printf("error signing transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	xdrTxBase64, err := signedTx.Base64()
	if err != nil {
		log.Printf("error encoding transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	fmt.Println(xdrTxBase64)
	res, err := sorobanClient.SubmitTransactionXDR(ctx, xdrTxBase64)
	if err != nil {
		log.Printf("error submitting transaction: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSubmitting,
			Headers:    headers,
		}
	}
	log.Printf("revoked sponsorship result: %v", res.ResultXdr)
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       res.Hash,
		Headers:    headers,
	}
}

func handleListRequest(ctx context.Context, req events.LambdaFunctionURLRequest) events.LambdaFunctionURLResponse {
	limit := req.QueryStringParameters["limit"]
	if limit == "" {
		limit = "100"
	}
	order := req.QueryStringParameters["order"]
	if order == "" {
		order = "asc"
	}
	accounts, err := sorobanClient.GetCreateAccountOperation(ctx, key.Address(), limit, order)
	if err != nil {
		log.Printf("error getting account operations: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	if len(accounts) == 0 {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       ErrorNoSponsorship,
			Headers:    headers,
		}
	}

	var accountInfos []listAccountResponse

	for _, account := range accounts {
		var balance string
		for _, b := range account.Balance {
			if b.AssetType == "native" {
				balance = b.Balance
			}
		}
		accountInfos = append(accountInfos, listAccountResponse{account.AccountID, balance})
	}

	res, err := json.Marshal(accountInfos)
	if err != nil {
		log.Printf("error marshalling accounts: %v", err)
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       ErrorSponsorship,
			Headers:    headers,
		}
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       string(res),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}
