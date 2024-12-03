package soroban

import (
	"github.com/stellar/go/keypair"
)

func LoadKeystore(seed string) (*keypair.Full, error) {
	fullKeyPair, err := keypair.ParseFull(seed)
	if err != nil {
		return nil, err
	}
	return fullKeyPair, nil
}
