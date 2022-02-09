package common

import (
	"github.com/boltdb/bolt"
	"log"
)

const (
	DBFile = "blockchain.db"
	BlocksBucket = "blocks"
	PeerBucket = "peerBucket"
)


var (
	// TempBlocks 存放临时区块
	TempBlocks []Block
	// CandidateBlokcs 候选区块的通道
	CandidateBlokcs = make(chan Block)
	// Announcements 广播通道
	Announcements = make(chan string)
	// AddStatus 通知关闭peer的通道
	AddStatus = make(chan bool)
)


type Blockchain struct {
	LastBlockHash string
	DB *bolt.DB
}

// Delegate 验证者结构
type Delegate struct {
	Address string
	Number int
}

// Block 区块结构
type Block struct {
	// 高度
	Height int64
	// 时间戳
	Timestamp string
	// 数据
	Data string
	// 哈希
	Hash string
	// 前区块哈希
	PrevHash string
	// 验证者
	Validator string
}

// GetDB 获取数据库句柄
func GetDB() *bolt.DB {
	db, err := bolt.Open(DBFile, 0600, nil)
	if err != nil{
		log.Fatal(err)
	}
	return db
}
