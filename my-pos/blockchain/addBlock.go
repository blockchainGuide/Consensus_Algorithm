package blockchain

import (
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-pos/common"
	"my-pos/consensus"
	"time"
)

var addBlockCmd = &cobra.Command{
	Use: "addBlock",
	Short: "addBlock",
	Long: "add a block to blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := cmd.Flags().GetString("data")
		if err != nil{
			log.Fatal("please input a data")
		}
		addBlock(data)
	},
}

func init()  {
	addBlockCmd.Flags().StringP("data", "d", "", "data")
}

func addBlock(data string)  {
	// 把客户端输入的数据封装成区块
	newBlock := generateBlock(data)

	// 开始共识
	go consensus.Start()

	// 把新生成的区块添加到候选区块通道中
	common.CandidateBlokcs <- newBlock

	select {
	// 阻塞等待退出
	case <- common.ExitChan:
		return
	}
}

func generateBlock(data string) common.Block  {

	var lastBlock common.Block
	// 获取最后一个区块
	db := common.GetDB()
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(common.BlocksBucket))
		lastHash := bucket.Get([]byte("lastHash"))
		lastBlockBytes := bucket.Get(lastHash)
		lastBlock = deserializeBlock(lastBlockBytes)
		return nil
	})
	var newBlock common.Block
	newBlock.Index = lastBlock.Index + 1
	newBlock.TimeStamp = time.Now().String()
	newBlock.Data = data
	newBlock.PrevHash = lastBlock.HashCode
	newBlock.Validator = common.Validator{}
	newBlock.HashCode = GenerateHashValue(newBlock)
	return newBlock
}
