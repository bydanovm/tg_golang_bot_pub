package database

import "time"

type Cryptoprices struct {
	CryptoId     int
	CryptoPrice  float32
	CryptoUpdate time.Time
}
