# Stellar sponsored account

## API

Base URL: `https://ciihnqaqiomjdoicuy5rgwmy5m0vxanz.lambda-url.us-east-1.on.aws/`
Public key: `GCV5PJ4H57MZFRH5GM3E3CNFLWQURNFNIHQOYGRQ7JHGWJLAR2SFVZO6`

### Sign transaction

Sign the sponsored account transaction with the wallet you want to be sponsored.

```bash
curl -X POST -H 'Content-Type: application/json' $BASE_URL -d '{"data": "AAA..."}'
```

Response:

```text
Status: 200 OK
Hash: 0x7b2...
```
