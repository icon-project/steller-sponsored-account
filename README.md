# Stellar sponsored account

This is a Lambda function that signs a transaction with a sponsored account and submits it to the Stellar network.

## Usage

Base URL: `https://ciihnqaqiomjdoicuy5rgwmy5m0vxanz.lambda-url.us-east-1.on.aws/`

Public key: `GCV5PJ4H57MZFRH5GM3E3CNFLWQURNFNIHQOYGRQ7JHGWJLAR2SFVZO6`

### Sign transaction

Sign the sponsored account transaction with the wallet you want to be sponsored.

Send a POST request with the XDR base64 encoded transaction data.

```bash
curl -X POST -H 'Content-Type: application/json' $BASE_URL -d '{"data": "AAA..."}'
```

Response:

```text
Status: 200 OK
Context-Type: text/plain
Body: 0x7b2...
```
