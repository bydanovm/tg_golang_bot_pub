package database

import (
	"fmt"
	"time"
)

type Users struct {
	Id        int       `sql_type:"SERIAL PRIMARY KEY"`
	Timestamp time.Time `sql_type:"TIMESTAMP DEFAULT CURRENT_TIMESTAMP"`
	UserName  string    `sql_type:"INT"`
	Chat_Id   int       `sql_type:"TEXT"`
	Message   string    `sql_type:"NUMERIC(15,3)"`
	Answer    string    `sql_type:"TIMESTAMP"`
}

type DictCrypto struct {
	Id              int       `sql_type:"SERIAL PRIMARY KEY"`
	Timestamp       time.Time `sql_type:"TIMESTAMP DEFAULT CURRENT_TIMESTAMP"`
	CryptoId        int       `sql_type:"INT"`
	CryptoName      string    `sql_type:"TEXT"`
	CryptoLastPrice float32   `sql_type:"NUMERIC(15,3)"`
	CryptoUpdate    time.Time `sql_type:"TIMESTAMP"`
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
	Id           int       `sql_type:"SERIAL PRIMARY KEY"`
	Timestamp    time.Time `sql_type:"TIMESTAMP DEFAULT CURRENT_TIMESTAMP"`
	CryptoId     int       `sql_type:"INT"`
	CryptoPrice  float32   `sql_type:"NUMERIC(15,3)"`
	CryptoUpdate time.Time `sql_type:"TIMESTAMP"`
}

type Expressions struct {
	Key      string
	Operator string
	Value    string
}

func (exp *Expressions) Join() string {
	return fmt.Sprintf("%s %s %s AND ", exp.Key, exp.Operator, exp.Value)
}
