package database

import (
	"fmt"
	"time"
)

type Users struct {
	Id        int
	Timestamp time.Time
	UserName  string
	Chat_Id   int
	Message   string
	Answer    string
}

type DictCrypto struct {
	Id              int
	Timestamp       time.Time
	CryptoId        int
	CryptoName      string
	CryptoLastPrice float32
	CryptoUpdate    time.Time
}

const (
	EQ              string = "="
	Id              string = "id"
	Timestamp       string = "timestamp"
	CryptoId        string = "cryptoid"
	CryptoName      string = "cryptoname"
	CryptoLastPrice string = "cryptolastorice"
	CryptoUpdate    string = "cryptoupdate"
)

// Структура данных таблицы Cryptoprices
type Cryptoprices struct {
	Id           int
	Timestamp    time.Time
	CryptoId     int
	CryptoPrice  float32
	CryptoUpdate time.Time
}

type Expressions struct {
	Key      string
	Operator string
	Value    string
}

func (exp *Expressions) Join() string {
	return fmt.Sprintf("%s %s %s AND ", exp.Key, exp.Operator, exp.Value)
}

type ReturnValues struct {
	Keys   string
	Values interface{}
}

func (rv *ReturnValues) Scan() {

}
