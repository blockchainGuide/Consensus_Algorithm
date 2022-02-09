package consensus

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"log"
	"my-pos/common"
)

var viewValidatorCmd = &cobra.Command{
	Use: "view",
	Short: "view",
	Long: "view Validator",
	Run: func(cmd *cobra.Command, args []string) {
		view()
	},
}

func view()  {
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

	spew.Dump(validators)
}

func deserializeValidator (encoderValidator []byte) common.Validator {
	var validator common.Validator
	err := json.Unmarshal(encoderValidator, &validator)
	if err != nil{
		log.Fatal("unmarshal block error: ", err)
	}
	return validator
}


