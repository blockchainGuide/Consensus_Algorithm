package consensus

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"my-pos/common"
	"sync"
	"time"
)

var mutex sync.Mutex


func ConsensusCmd () *cobra.Command{
	consensusCmd.AddCommand(addValidatorCmd)
	consensusCmd.AddCommand(viewValidatorCmd)
	consensusCmd.AddCommand(initValidatorCmd)
	return consensusCmd
}
var consensusCmd = &cobra.Command{
	Use: "consensus",
	Short: "consensus",
	Long: "consensus manage",
	Run: func(cmd *cobra.Command, args []string) {
	},
}


func Start(){
	fmt.Println("开始共识...")
	// 不断读取候选区块通道，如果有新的区块，就追加到临时区块切片中
	go func() {
		for candidate := range common.CandidateBlokcs{
			fmt.Println("有新的临时区块")
			common.TempBlocks = append(common.TempBlocks, candidate)
		}
	}()


	for  {
		// 循环pos共识算法
		pos()
	}

}

func pos()  {
	// 复制临时区块
	temp := common.TempBlocks

	// 根据temp的长度判断是否存在临时区块
	if len(temp) > 0{
		fmt.Println("准备出块...")
		var tempValidators []common.Validator
		// 获取所有的验证者
		validators := getAllValidators()
		// 根据验证者拥有的token数量及时间得出权重，权重越高，被选取为出块节点的概率越大
		for i := 0; i < len(validators) - 1; i++{
			for j := 0; j < validators[i].Tokens * validators[i].Days; j++{
				tempValidators = append(tempValidators, validators[i])
			}
		}

		// 获取数据库句柄
		db := common.GetDB()
		defer db.Close()
		for _, block := range temp{
			// 挑选验证者
			rand.Seed(time.Now().Unix())
			var rd =rand.Intn(len(tempValidators))
			block.Validator = validators[rd %len(validators)]
			// 持久化
			db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(common.BlocksBucket))
				err := bucket.Put([]byte(block.HashCode), serializeBlock(block))
				if err != nil {
					log.Fatal(err)
				}
				err = bucket.Put([]byte("lastHash"), []byte(block.HashCode))
				if err != nil {
					log.Fatal(err)
				}
				return nil
			})
			fmt.Println("区块添加到区块链完成！")
		}
		mutex.Lock()
		common.ExitChan <- true
		temp = []common.Block{}
		common.TempBlocks = []common.Block{}
		mutex.Unlock()
	}
}

func getAllValidators() []common.Validator {
	var validators []common.Validator

	db := common.GetDB()
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next(){
			if string(k) == "lastHash"{
				continue
			}
			block := deserializeValidator(v)
			validators = append(validators, block)
		}
		return nil
	})
	return validators
}

func serializeBlock (block common.Block)[]byte  {
	result, err := json.Marshal(block)
	if err != nil{
		log.Fatal("marshal block error: ", err)
	}
	return result
}
