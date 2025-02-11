package soroban

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/stellar/go/keypair"
)

const (
	jsonRPCVersion = "2.0"
	successStatus  = "SUCCESS"
	failedStatus   = "FAILED"
)

type Client struct {
	idCounter uint64
	http      *http.Client
	rpcUrl    string
	httpUrl   string
}

func New(rpcURL, httpURL string) (*Client, error) {
	return &Client{
		http:    &http.Client{},
		rpcUrl:  rpcURL,
		httpUrl: httpURL,
	}, nil
}

func (c *Client) CallContext(ctx context.Context, result interface{}, method string, params interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}

	msg, err := c.newMessage(method, params)
	if err != nil {
		return err
	}

	respBody, err := c.doRequest(ctx, msg)
	if err != nil {
		return err
	}
	defer respBody.Close()

	var respmsg RPCResponse
	if err := json.NewDecoder(respBody).Decode(&respmsg); err != nil {
		return err
	}
	if respmsg.Error != nil {
		return respmsg.Error
	}
	if len(respmsg.Result) == 0 {
		return fmt.Errorf("result is empty")
	}

	return json.Unmarshal(respmsg.Result, result)
}

func (c *Client) newMessage(method string, paramsIn interface{}) (*RPCRequest, error) {
	msg := &RPCRequest{Version: jsonRPCVersion, ID: c.nextID(), Method: method}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		if msg.Params, err = json.Marshal(paramsIn); err != nil {
			return nil, err
		}
	}
	return msg, nil
}

func (c *Client) doRequest(ctx context.Context, msg interface{}) (io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcUrl, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	req.Header.Set("Content-Type", "application/json")

	// do request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		var body []byte
		if _, err := buf.ReadFrom(resp.Body); err == nil {
			body = buf.Bytes()
		}

		return nil, HTTPError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return resp.Body, nil
}

func (c *Client) nextID() json.RawMessage {
	id := atomic.AddUint64(&c.idCounter, 1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

func (c *Client) GetTransaction(ctx context.Context, txHash string) (*TransactionResponse, error) {
	txn := &TransactionResponse{}
	params := make(map[string]string)
	params["hash"] = txHash
	if err := c.CallContext(ctx, txn, "getTransaction", params); err != nil {
		return nil, err
	}
	return txn, nil
}

func (c *Client) SubmitTransactionXDR(ctx context.Context, txXDR string) (*TransactionResponse, error) {
	txn := &TxnCreationResponse{}
	params := make(map[string]string)
	params["transaction"] = txXDR
	if err := c.CallContext(ctx, txn, "sendTransaction", params); err != nil {
		return nil, err
	}
	log.Printf("txn hash: %s", txn.Hash)
	return c.waitForSuccess(ctx, txn.Hash)
}

func (c *Client) waitForSuccess(ctx context.Context, txHash string) (*TransactionResponse, error) {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction success")
		case <-ticker.C:
			txn, err := c.GetTransaction(ctx, txHash)
			if err != nil {
				return nil, err
			}
			switch txn.Status {
			case successStatus:
				return txn, nil
			case failedStatus:
				return nil, fmt.Errorf("transaction failed: %s", txn.ResultXdr)
			}
		}
	}
}

func (c *Client) GetNetworkInfo() (*NetworkInfo, error) {
	network := &NetworkInfo{}
	if err := c.CallContext(context.Background(), network, "getNetwork", nil); err != nil {
		return nil, err
	}
	return network, nil
}

func (c *Client) LoadKeystore(seed string) *keypair.Full {
	return keypair.MustParseFull(seed)
}

func (c *Client) GetAccount(ctx context.Context, address string) (*AccountInfo, error) {
	account := &AccountInfo{}
	res, err := c.http.Get(c.httpUrl + "/accounts/" + address)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Client) GetCreateAccountOperation(ctx context.Context, account, limit, order string) ([]AccountInfo, error) {
	ops := struct {
		Embedded struct {
			Records []AccountSponsored `json:"records"`
		} `json:"_embedded"`
	}{}

	url := fmt.Sprintf("%s/accounts?sponsor=%s&limit=%s&order=%s&include_failed=false", c.httpUrl, account, limit, order)
	res, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&ops); err != nil {
		return nil, err
	}
	var accounts []AccountInfo
	for _, op := range ops.Embedded.Records {
		if op.Sponsor == account {
			accounts = append(accounts, AccountInfo{
				AccountID: op.AccountID,
				Sponser:   op.Sponsor,
				Balance:   op.Balances,
			})
		}
	}
	return accounts, nil
}
