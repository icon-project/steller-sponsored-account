package soroban

import "github.com/stellar/go/keypair"

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

type AccountInfo struct {
	AccountID string `json:"account_id"`
	Sequence  string `json:"sequence"`
}

type ExecuteSponsoredRequest struct {
	NetworkPassphrase string
	Address           string
	Key               *keypair.Full
	Data              string
}
