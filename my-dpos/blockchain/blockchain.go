package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-dpos/common"
	"time"
)

func BlockchainCmd() *cobra.Command {
	blockchainCmd.AddCommand(createCmd)
	blockchainCmd.AddCommand(blockCmd)
	return blockchainCmd
}

// blockchainCmd represents the blockchain command
var blockchainCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "blockchan",
	Long: `blockchain manage`,
}

// genGenesisBlock 生成创世区块
func genGenesisBlock() common.Block{
	block := common.Block{0,time.Now().String(),"","","","root"}
	// 计算区块哈希
	block.Hash = calculateBlockHash(block)
	return block
}

// calculateBlockHash 计算区块的哈希
func calculateBlockHash(block common.Block) string{
	record := string(block.Height)+ block.Timestamp + block.Data + block.PrevHash
	return calculateHash(record)
}

// GetLastHeight 返回最高区块的高度
func getLastHeight() int64  {
	var lastBlock common.Block
	db := common.GetDB()
	defer db.Close()
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.BlocksBucket))
		lastHash := bucket.Get([]byte("lastHash"))
		blockData := bucket.Get(lastHash)
		lastBlock = deserializeBlock(blockData)
		return nil
	})
	if err != nil{
		log.Panic("get last block hash error: ", err)
	}
	return lastBlock.Height
}

// getLastBlockHash 获取最后一个区块的哈希
func getLastBlockHash() string  {
	var lastBlock common.Block
	db := common.GetDB()
	defer db.Close()
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.BlocksBucket))
		lastHash := bucket.Get([]byte("lastHash"))
		blockData := bucket.Get(lastHash)
		lastBlock = deserializeBlock(blockData)
		return nil
	})
	if err != nil{
		log.Panic(err)
	}
	return lastBlock.Hash
}

// SetupDB 初始化数据库
func setupDB() () {
	db, err := bolt.Open(common.DBFile, 0600, nil)
	defer db.Close()
	if err != nil{
		log.Println(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_,err := tx.CreateBucketIfNotExists([]byte(common.BlocksBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
	err = db.Update(func(tx *bolt.Tx) error {
		_,err := tx.CreateBucketIfNotExists([]byte(common.PeerBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
}

// SerializeBlock 区块序列化
func serializeBlock (block common.Block)[]byte  {
	result, err := json.Marshal(block)
	if err != nil{
		log.Fatal("marshal block error: ", err)
	}
	return result
}

// DeserializeBlock 区块反序列化
func deserializeBlock (encoderBlock []byte) common.Block {
	var block common.Block
	err := json.Unmarshal(encoderBlock, &block)
	if err != nil{
		log.Fatal("unmarshal block error: ", err)
	}
	return block
}




