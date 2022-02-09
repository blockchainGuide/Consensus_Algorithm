package blockchain

import (
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-dpos/common"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create",
	Long:  `create blockchain`,
	Run: func(cmd *cobra.Command, args []string) {
		NewBlockchain()
	},
}

// NewBlockchain 创建一个区块链
func NewBlockchain()  {
	// 初始化数据库
	setupDB()
	db := common.GetDB()
	defer db.Close()

	// 获取最后一个区块的哈希
	var lastBlockHash []byte
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.BlocksBucket))
		lastBlockHash = bucket.Get([]byte("lastHash"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// lastBlockHash等于0，没有区块，存放创世区块
	if len(lastBlockHash) == 0{
		// 生成创世区块
		block := genGenesisBlock()
		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(common.BlocksBucket))
			err = bucket.Put([]byte("lastHash"), []byte(block.Hash))
			err = bucket.Put([]byte(block.Hash), serializeBlock(block))
			if err != nil {
				log.Panic(err)
			}
			return nil
		})
		//common.Announcements <- fmt.Sprintf("生成创世块成功")
	}
	log.Println(">>> 创建区块链成功")
}

