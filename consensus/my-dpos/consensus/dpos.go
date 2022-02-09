package consensus

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"my-dpos/common"
	"time"
)

func DPoSCmd() *cobra.Command  {
	dposCmd.AddCommand(voteCmd)
	return dposCmd
}
var dposCmd = &cobra.Command{
	Use: "dpos",
	Short: "consensus",
	Long: "consensus manage",

}
// Start 开启共识
func Start() {

	// 循环读取候选区块通道中的区块信息
	go func() {
		for{
			for candidate := range common.CandidateBlokcs{
				common.TempBlocks = append(common.TempBlocks, candidate)
			}
		}
	}()
	// 开始共识
	for  {
		if len(common.TempBlocks) > 0{
			for _, block := range common.TempBlocks{
				// 挑选验证者
				delegate := getDelegate()
				block.Validator = delegate.Address
				// 序列化
				result, err := json.Marshal(block)
				if err != nil{
					log.Fatal("marshal block error: ", err)
				}
				//添加到数据库
				db := common.GetDB()

				db.Update(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(common.BlocksBucket))
					bucket.Put([]byte(block.Hash), result)
					err = bucket.Put([]byte("lastHash"),[]byte(block.Hash))
					return nil
				})
				log.Println(">>> 添加区块成功！")
				db.Close()

				common.TempBlocks = nil
				// todo
				// 把添加的区块广播出去
				//common.Announcements <- fmt.Sprintf("新的区块生成, 高度为：%d", block.Height)
			}
		}

	}
}

// getDelegate 获取代理人
func getDelegate() common.Delegate {
	// 从数据库中查出所有的候选者
	var delegates []common.Delegate

	db := common.GetDB()
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			delegate := unmarshalDelegate(v)
			delegates = append(delegates, delegate)
		}
		return nil
	})

	// 挑选验证者，假设取得票数最高的两个节点作为验证者
	quickSort(0, len(delegates) - 1, delegates)
	rand.Seed(time.Now().Unix())
	randNumber := rand.Intn(2)

	// 返回指定的验证者
	return delegates[randNumber]
}

// quickSort 快速排序
func quickSort(left int, right int, array []common.Delegate) {
	l := left
	r := right

	pivot := array[(left+right)/2].Number


	for l < r {

		for array[l].Number < pivot {
			l++
		}

		for array[r].Number > pivot {
			r--
		}

		if l >= r {
			break
		}

		temp := array[l]
		array[l] = array[r]
		array[r] = temp

		if array[l].Number == pivot {
			r--
		}
		if array[r].Number == pivot {
			l++
		}
	}

	if l == r {
		l++
		r--
	}

	if left < r {
		quickSort(left, r, array)
	}

	if right > l {
		quickSort(l, right, array)
	}
}

func unmarshalDelegate(marshalDelegate []byte) common.Delegate {
	var delegate common.Delegate
	err := json.Unmarshal(marshalDelegate, &delegate)
	if err != nil{
		log.Fatal("unmarshal delegate error: ", err)
	}
	return delegate
}
