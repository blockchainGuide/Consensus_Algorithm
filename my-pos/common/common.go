package common

import (
	"github.com/boltdb/bolt"
	"log"
	"sync"
)

const (
	DBFile = "blockchain.db"
	BlocksBucket = "blocksBucket"
	PeerBucket = "peersBucket"
)

var TempBlocks []Block

var CandidateBlokcs = make(chan Block)
var Validators = make(map[string]int)
var ExitChan = make(chan bool)

var Mutex sync.Mutex

type Block struct {
	Index     int
	TimeStamp string
	Data       string
	HashCode  string
	PrevHash  string
	Validator Validator
}

type Validator struct {
	Tokens int
	Days int
	Address string
}

func GetDB() *bolt.DB {
	db, err := bolt.Open(DBFile, 0600, nil)
	if err != nil{
		log.Fatal(err)
		return nil
	}
	return db
}
