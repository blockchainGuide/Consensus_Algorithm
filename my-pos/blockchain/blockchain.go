package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"log"
	"my-pos/common"
	"strconv"
)

type Blockchain struct {
	DB *bolt.DB
}

func BlockchianCmd() *cobra.Command {
	blockchainCmd.AddCommand(viewBlockchainCmd)
	blockchainCmd.AddCommand(addBlockCmd)
	blockchainCmd.AddCommand(createBlockchainCmd)
	return blockchainCmd
}

var blockchainCmd = &cobra.Command{
	Use: "blockchain",
	Short: "blockchain",
	Long: "block chain manage",
	Run: func(cmd *cobra.Command, args []string) {
	},
}



func GenerateHashValue(block common.Block) string  {
	result, err := json.Marshal(block.Validator)
	if err != nil{
		log.Fatal(err)
	}
	var hashcode = block.PrevHash + block.TimeStamp + string(result) + block.Data + strconv.Itoa(block.Index)
	return calculateHash(hashcode)
}

func calculateHash(s string) string  {
	var sha = sha256.New()
	sha.Write([]byte(s))
	hashed := sha.Sum(nil)
	return hex.EncodeToString(hashed)
}

func serializeBlock (block common.Block)[]byte  {
	result, err := json.Marshal(block)
	if err != nil{
		log.Fatal("marshal block error: ", err)
	}
	return result
}

func deserializeBlock (encoderBlock []byte) common.Block {
	var block common.Block
	err := json.Unmarshal(encoderBlock, &block)
	if err != nil{
		log.Fatal("unmarshal block error: ", err)
	}
	return block
}