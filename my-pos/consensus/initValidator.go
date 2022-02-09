package consensus

import (
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-pos/common"
)

var initValidatorCmd = &cobra.Command{
	Use: "init",
	Short: "init",
	Long: "init validator",
	Run: func(cmd *cobra.Command, args []string) {
		initValidator()
	},
}

func initValidator(){
	db := common.GetDB()
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(common.PeerBucket))
		if err != nil{
			log.Fatal(err)
			return err
		}
		return nil
	})
}
