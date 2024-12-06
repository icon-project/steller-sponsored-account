package soroban

import (
	"encoding/json"
	"fmt"
)

type TransactionResponse struct {
	Status                string `json:"status"`
	LatestLedger          int64  `json:"latestLedger"`
	LatestLedgerCloseTime string `json:"latestLedgerCloseTime"`
	OldestLedger          int64  `json:"oldestLedger"`
	OldestLedgerCloseTime string `json:"oldestLedgerCloseTime"`
	ApplicationOrder      int64  `json:"applicationOrder"`
	EnvelopeXdr           string `json:"envelopeXdr"`
	ResultXdr             string `json:"resultXdr"`
	ResultMetaXdr         string `json:"resultMetaXdr"`
	Ledger                int64  `json:"ledger"`
	CreatedAt             string `json:"createdAt"`
	Hash                  string
}

type TxnCreationResponse struct {
	Status                string `json:"status"`
	Hash                  string `json:"hash"`
	LatestLedger          int64  `json:"latestLedger"`
	LatestLedgerCloseTime string `json:"latestLedgerCloseTime"`
}

type NetworkInfo struct {
	Passphrase string `json:"passphrase"`
}

type RPCRequest struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type RPCResponse struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (err *RPCError) Error() string {
	return fmt.Sprintf("json-rpc error with code: %d, message: %s, & data: %v", err.Code, err.Message, err.Data)
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (err HTTPError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s", err.Status, err.Body)
}
