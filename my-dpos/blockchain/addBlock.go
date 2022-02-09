package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/spf13/cobra"
	"log"
	"my-dpos/common"
	"my-dpos/consensus"
	"my-dpos/p2p"
	"time"
)

var blockCmd = &cobra.Command{
	Use: "addBlock",
	Short: "add",
	Long: "add a block to blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := cmd.Flags().GetString("data")
		if err != nil{
			log.Fatal("please input a data")
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil{
			log.Fatal(err)
		}

		target, err := cmd.Flags().GetString("target")
		if err != nil{
			log.Fatal(err)
		}

		add(data, port, target)
	},
}

func init()  {
	blockCmd.Flags().StringP("data", "d", "", "data")
	blockCmd.Flags().IntP("port", "p", 0, "port")
	blockCmd.Flags().StringP("target", "t", "", "target")
}

// add 添加区块
func add(data string, port int, target string)  {
	log.Println(">>> 开始添加区块链")
	// 启动共识
	go consensus.Start()
	log.Println(">>> 启动共识...")

	// 启动p2p节点
	go p2p.Start(port, target, 0)
	log.Println(">>> 启动p2p网络...")


	// 打包区块
	newBlock := common.Block{}
	newBlock.Height = getLastHeight() + 1
	newBlock.Data = data
	newBlock.Timestamp = time.Now().String()
	newBlock.PrevHash = getLastBlockHash()
	newBlock.Hash = calculateBlockHash(newBlock)

	// 新生成的区块添加到候选区块通道中，等待共识确认
	common.CandidateBlokcs <- newBlock


	for{

	}
	//fmt.Println("等待关闭peer节点........")
	//select {
	//	case flag := <- common.AddStatus:
	//		fmt.Println(flag)
	//		if flag {
	//			return
	//		}
	//}

}


// calculateHash 计算哈希
func calculateHash(s string) string{
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}


