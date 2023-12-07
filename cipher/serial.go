package cipher

import (
	"crypto/rand"
	log "github.com/sirupsen/logrus"
	"math/big"
)

type SerialNumberGenerator interface {
	Generate() (*big.Int, error)
}

type randSerialNum struct{}

func (r *randSerialNum) Generate() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		log.Errorln("generate serial number failed:", err)
		return nil, err
	}

	return sn, nil
}
