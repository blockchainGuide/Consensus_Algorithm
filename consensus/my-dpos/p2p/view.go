package p2p

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"log"
	"my-dpos/common"
)

var viewCmd = &cobra.Command{
	Use: "view",
	Short: "view peer",
	Long: "view all peer in network",
	Run: func(cmd *cobra.Command, args []string) {
		view()
	},
}

// view 查看所有在网络中的节点
func view()  {
	var delegates []common.Delegate

	db := common.GetDB()
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next(){
			delegate := unmarshalDelegate(v)
			delegates = append(delegates, delegate)
		}
		return nil
	})
	spew.Dump(delegates)
}

func unmarshalDelegate(marshalDelegate []byte) common.Delegate {
	var delegate common.Delegate
	err := json.Unmarshal(marshalDelegate, &delegate)
	if err != nil{
		log.Fatal("unmarshal delegate error: ", err)
	}
	return delegate
}