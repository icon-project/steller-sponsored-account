# Stellar sponsored account

## API

### Create Sponsored Account

Base URL: `https://ciihnqaqiomjdoicuy5rgwmy5m0vxanz.lambda-url.us-east-1.on.aws/`

### Sign transaction

Grab the signed XDR transaction from the response and sign it with the account and submit.

```bash
curl -X POST -H 'Content-Type: application/json' $BASE_URL -d '{"data": "AAA..."}'
```

Response:

```text
Status: 200 OK
Hash: 0x7b2...
```
