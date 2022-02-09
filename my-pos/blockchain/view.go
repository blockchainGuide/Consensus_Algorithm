package blockchain

import (
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"my-pos/common"
)

var viewBlockchainCmd = &cobra.Command{
	Use: "view",
	Short: "view",
	Long: "view block chain",
	Run: func(cmd *cobra.Command, args []string) {
		view()
	},
}

func view()  {

	var blocks []common.Block

	db := common.GetDB()
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.BlocksBucket))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next(){
			if string(k) == "lastHash"{
				continue
			}
			block := deserializeBlock(v)
			blocks = append(blocks, block)
		}
		return nil
	})

	spew.Dump(blocks)
}


