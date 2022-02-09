package consensus

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-dpos/common"
)

var voteCmd = &cobra.Command{
	Use: "vote",
	Short: "vote",
	Long: "vote to a address",
	Run: func(cmd *cobra.Command, args []string) {

		address, err := cmd.Flags().GetString("address")
		if err != nil{
			log.Fatal(err)
		}
		number, err := cmd.Flags().GetInt("number")
		if err != nil{
			log.Fatal(err)
		}

		Vote(address, number)
	},
}

func init()  {
	voteCmd.Flags().IntP("number", "n", 0, "number")
	voteCmd.Flags().StringP("address", "a", "", "address")
}
func Vote(address string, voteNum int)  {
	if address == "" {
		log.Fatal("节点名称不能为空")
	}
	if voteNum < 0 {
		log.Fatal("最小投票数为1")
	}
	var delegate common.Delegate
	db := common.GetDB()
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		result := bucket.Get([]byte(address))
		err := json.Unmarshal(result, &delegate)
		if err != nil{
			log.Fatal("unmarshal delegate error: ", err)
		}
		log.Println(delegate)
		delegate.Number += voteNum
		mashalResult, err := json.Marshal(delegate)
		if err != nil{
			log.Fatal("mashal delegate error: ", err)
		}
		err = bucket.Put([]byte(address), mashalResult)
		if err != nil{
			log.Fatal("put delegate error: ", err)
		}
		return nil
	})
}
