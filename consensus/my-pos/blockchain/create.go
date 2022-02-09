package blockchain

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"my-pos/common"
	"time"
)

var createBlockchainCmd = &cobra.Command{
	Use: "create",
	Short: "create",
	Long: "create blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		create()
	},
}

func create() {
	// 生成创世块
	genesisBlock := common.Block{}
	genesisBlock = common.Block{
		0,
		time.Now().String(),
		"",
		GenerateHashValue(genesisBlock),
		"",
		common.Validator{},
	}

	// 获取bolt数据库句柄
	db := common.GetDB()
	defer db.Close()
	// 把创世块添加进数据库文件中
	db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(common.BlocksBucket))
		if err != nil{
			fmt.Println(err)
			return err
		}

		// 获取最后一个区块的哈希
		lastHash := bucket.Get([]byte("lastHash"))
		if lastHash == nil {
			bucket.Put([]byte("lastHash"), []byte(genesisBlock.HashCode))
			bucket.Put([]byte(genesisBlock.HashCode), serializeBlock(genesisBlock))
		}
		return err
	})
}
