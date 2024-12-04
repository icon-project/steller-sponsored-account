# Stellar sponsored account

## API

### Create Sponsored Account

Base URL: `https://ciihnqaqiomjdoicuy5rgwmy5m0vxanz.lambda-url.us-east-1.on.aws/`

```bash
curl -X POST $BASE_URL -H 'Content-Type: application/json' -d '{"address": "GCFAPEJDCDCYSZAFTRWD2L2X3ARKUWJH7LVDN5ZWSLGF6NOQNWVNIR2X"}'
```

Response:

```text
AAAAAgAAAACKB5EjEMWJZAWcbD0vV9giqlkn+uo29zaSzF810G2q1AAAAAAAE4kfAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAAABAAAAAIoHkSMQxYlkBZxsPS9X2CKqWSf66jb3NpLMXzXQbarUAAAAEAAAAACKB5EjEMWJZAWcbD0vV9giqlkn+uo29zaSzF810G2q1AAAAAAAAAAAAAAAAIoHkSMQxYlkBZxsPS9X2CKqWSf66jb3NpLMXzXQbarUAAAAAAAAAAAAAAABAAAAAIoHkSMQxYlkBZxsPS9X2CKqWSf66jb3NpLMXzXQbarUAAAAEQAAAAAAAAAB0G2q1AAAAEDlth/eKuCmdtDWTllTKdHR591a+yaaF95ZAnIdS25ZdqDEkIYjZW757sFekR+0O4e/xrNr414klY/yUMpoARoI
```

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
