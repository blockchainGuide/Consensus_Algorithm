package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	net "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"io"
	"log"
	mrand "math/rand"
	"my-dpos/common"
)

func P2PCmd() *cobra.Command  {
	p2pCmd.AddCommand(viewCmd)
	return p2pCmd
}

var p2pCmd = &cobra.Command{
	Use: "p2p",
	Short: "p2p manager",
	Long: "p2p manager",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil{
			log.Fatal(err)
		}
		seed, err := cmd.Flags().GetInt64("seed")
		if err != nil{
			log.Fatal(err)
		}

		target, err := cmd.Flags().GetString("target")
		if err != nil{
			log.Fatal(err)
		}

		Start(port, target, seed)

	},
}

func init()  {
	p2pCmd.Flags().IntP("port", "p", 0, "port")
	p2pCmd.Flags().Int64P("seed", "s", 0, "seed")
	p2pCmd.Flags().StringP("target", "t", "", "target")
}


// Start 开启p2p网络
func Start(port int, target string, seed int64) {

	// 创建一个libp2p节点
	ha, err := makeBasicHost(port, seed)
	if err != nil {
		log.Fatal(err)
	}

	// 判断是否有目标地址
	if target == "" {
		log.Println(">>> 等待新的连接")
		// 设置stream
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)
		select {}
	}else{

		// 设置stream
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		// 获取peer ID
		ipfsaddr, err := ma.NewMultiaddr(target)
		if err != nil {
			log.Fatalln(err)
		}
		info, err := peer.AddrInfoFromP2pAddr(ipfsaddr)
		if err != nil {
			log.Println(err)
		}

		// 把节点添加到peer store中，以便libp2p连接这个节点
		ha.Peerstore().AddAddr(info.ID, info.Addrs[0], peerstore.PermanentAddrTTL)

		// 创建当前节点与目标节点的stream
		s, err := ha.NewStream(context.Background(), info.ID, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}


		// 读写stream中的数据
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go writeData(rw)
		go readData(rw)

		select {}
	}
	return
}

func makeBasicHost(listenPort int, randseed int64) (host.Host, error){

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil{
		return nil, err
	}

	// libp2p2的配置
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	// 创建libp2p节点
	basicHost, err := libp2p.New(opts...)

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	// 保存候选人节点
	SavePeer(basicHost.ID().Pretty())
	log.Println(">>> 保存节点成功！")
	return basicHost, nil
}

func handleStream(s net.Stream)  {

	log.Println(">>> 获取一个新的连接")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	//添加成功
	for{
		var bc []*common.Block
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err.Error())
		}

		// 没有数据
		if str == ""{
			return
		}
		json.Unmarshal([]byte(str), &bc)

		bytes, err := json.MarshalIndent(bc, "", "")
		if err != nil{
			log.Println(">>> marshal indent bc error: ", err)
		}
		fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
	}
}

// writeData 将客户端数据处理写入BLockchain
func writeData(rw *bufio.ReadWriter) {

	// todo
	//log.Println(">>> 准备广播区块...")
	//var block *common.Block
	//var bc []*common.Block
	//for {
		//v, ok := <- common.Announcements
		//log.Println(v)
		//if ok {
		//	// 获取所有区块数据，转发出去
		//	db := common.GetDB()
		//	db.View(func(tx *bolt.Tx) error {
		//		bucket := tx.Bucket([]byte(common.BlocksBucket))
		//		cursor := bucket.Cursor()
		//		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
		//			if string(key) == "lastHash"{
		//				continue
		//			}
		//			block = deserializeBlock(value)
		//			bc = append(bc, block)
		//		}
		//		return nil
		//	})
		//	db.Close()

			//result, err := json.Marshal(bc)
			//if err != nil{
			//	log.Fatal("marshal blockchian error: ", err)
			//}
			//_, err = rw.WriteString(string(result)+"\n")
			//if err != nil{
			//	log.Println(">>> write data error: ", err)
			//}
			//rw.Flush()

			// 发生添加成功消息，等待关闭peer节点
			//common.AddStatus <- true
		//}
	//}
}

func SavePeer(name string)  {
	delegate := common.Delegate{name, 0}
	result, err := json.Marshal(delegate)
	if err != nil{
		log.Fatal("marshal delegate error: ", err)
	}
	db, err := bolt.Open(common.DBFile, 0600, nil)
	if err != nil{
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(common.PeerBucket))
		err = bucket.Put([]byte(name), result)
		if err != nil{
			log.Fatal(">>> save delegate error: ", err)
		}
		return nil
	})

}


// 区块反序列化
func deserializeBlock (encoderBlock []byte) *common.Block {
	var block *common.Block
	err := json.Unmarshal(encoderBlock, &block)
	if err != nil{
		log.Fatal("unmarshal block error: ", err)
	}
	return block
}