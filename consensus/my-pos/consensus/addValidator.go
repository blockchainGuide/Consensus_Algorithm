package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-pos/common"
	"time"
)

var addValidatorCmd = &cobra.Command{
	Use: "addV",
	Short: "add validator",
	Long: "add validator",
	Run: func(cmd *cobra.Command, args []string) {

		token, err := cmd.Flags().GetInt("token")
		if err != nil{
			log.Fatal(err)
		}

		addValidator(token)
	},
}

func init()  {
	addValidatorCmd.Flags().IntP("token", "t", 0, "token")
}
func addValidator(token int)  {
	// 验证者的地址用当前时间的sha256值代替
	address := calculateHash(time.Now().String())

	// 创建一个验证者
	validator := common.Validator{
		token,
		time.Now().Second(),
		address,
	}

	// 添加到数据库中
	db := common.GetDB()
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		err := bucket.Put([]byte(validator.Address), serializeValidator(validator))
		if err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	})
}

func calculateHash(s string) string  {
	var sha = sha256.New()
	sha.Write([]byte(s))
	hashed := sha.Sum(nil)
	return hex.EncodeToString(hashed)
}

func serializeValidator (validator common.Validator)[]byte  {
	result, err := json.Marshal(validator)
	if err != nil{
		log.Fatal("marshal block error: ", err)
	}
	return result
}
