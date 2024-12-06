package main

import (
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v6"
	"github.com/icon-project/steller-sponsored-account/soroban"
	"github.com/stellar/go/keypair"
)

type config struct {
	Soroban url.URL `env:"SOROBAN_RPC_URL,notEmpty"`
	Seed    string  `env:"KEY_SEED,notEmpty"`
}

var (
	cfg               config
	sorobanClient     *soroban.Client
	key               *keypair.Full
	networkPassphrase string
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	client, err := soroban.New(cfg.Soroban.String())
	if err != nil {
		log.Fatal(err)
	}
	networkInfo, err := client.GetNetworkInfo()
	if err != nil {
		log.Fatal(err)
	}
	networkPassphrase = networkInfo.Passphrase
	sorobanClient = client
	key = client.LoadKeystore(cfg.Seed)
}

func main() {
	lambda.Start(handleRequest)
}
