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
	Soroban url.URL `env:"SOROBAN_RPC,notEmpty"`
	Horizon url.URL `env:"HORIZON_URL,notEmpty"`
	Seed    string  `env:"SEED,notEmpty"`
}

var (
	cfg               config
	sorobanClient     *soroban.Client
	key               *keypair.Full
	NetworkPassphrase string
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	client, err := soroban.New(cfg.Soroban.String(), cfg.Horizon.String())
	if err != nil {
		log.Fatal(err)
	}
	networkInfo, err := client.GetNetworkInfo()
	if err != nil {
		log.Fatal(err)
	}
	NetworkPassphrase = networkInfo.Passphrase
	sorobanClient = client
	key, err = soroban.LoadKeystore(cfg.Seed)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	lambda.Start(handleRequest)
}
