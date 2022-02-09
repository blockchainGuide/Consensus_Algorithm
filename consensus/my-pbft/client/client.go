package client

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"log"
	"my-pbft/common"
	"sync"
	"time"
)

const ViewID = 0

var dataChan = make(chan []byte)
var sendDataChan = make(chan []byte)
var replyChan = make(chan string)

var protocolID string = "/chat/1.1.0"

func ClientCmd() *cobra.Command  {
	return clientCmd
}

var clientCmd = &cobra.Command{
	Use: "client",
	Short: "client manage",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取客户端的端口
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Println("get param error: ", err)
		}
		// 客户端传递的数据
		data, err := cmd.Flags().GetString("data")
		if err != nil{
			log.Println("get param error: ", err)
		}
		client := NewClient(port)
		client.Start(data)
	},
}

type Client struct {
	client host.Host
	// 密钥对
	keypair		Keypair
	// 所有的节点
	knownNodes	[]KnownNode
	mutex sync.Mutex
	// 响应日志
	replyLog	map[string]*common.ReplyMsg

}

type KnownNode struct {
	h peer.AddrInfo
}

type Keypair struct {
	// 私钥
	privkey			crypto.PrivKey
	// 公钥
	pubkey			crypto.PubKey
}

func init()  {
	clientCmd.Flags().IntP("port", "p", 0,"port")
	clientCmd.Flags().StringP("data", "d", "", "data")
}

// 创建客户端
func NewClient(listenPort int) *Client {
	// 生成密钥对
	r := rand.Reader
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, r)
	if err != nil{
		log.Println(err)
	}
	pubKey := prvKey.GetPublic()
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", "127.0.0.1", listenPort))

	// 创建libp2p节点
	h, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		 log.Println("创建的客户端节点失败：", err)
	}

	h.SetStreamHandler(protocol.ID(protocolID), handleStream)

	fmt.Printf(">>> 创建客户端p2p节点成功，客户端多路地址是: /ip4/%s/tcp/%v/p2p/%s\n", "0.0.0.0", listenPort, h.ID().Pretty())

	keyPair := Keypair{
		privkey: prvKey,
		pubkey: pubKey,
	}

	// 创建客户端
	client := &Client{
		h,
		keyPair,
		[]KnownNode{},
		sync.Mutex{},
		make(map[string]*common.ReplyMsg),
	}

	fmt.Println(">>> 创建客户端成功...")
	return client
}

func (c *Client) Start (data string)  {
	fmt.Println(">>> 开始启动客户端...")
	ctx := context.Background()
	// 通过协程获取网络中的节点，使用libp2p的mdns节点发现
	go c.getAllKonwons(c.client)
	// 发送客户端请求
	c.sendRequest(ctx, data)
	// 处理响应
	go c.handleConnection()

	select {}
}

func (c *Client) handleConnection(){
	// 读取传递的数据
	var reply []byte
	for{
		replyStr, ok := <- replyChan
		if ok{
			// 解析数据
			err := json.Unmarshal([]byte(replyStr), &reply)
			if err != nil{
				fmt.Println("解析回应错误...")
			}
			header, payload, _:= common.SplitMsg(reply)
			if err != nil {
				panic(err)
			}
			switch header {
			case common.HReply:
				c.handleReply(payload)
			}
		}
	}
}

func (c *Client) handleReply(payload []byte) {
	fmt.Println(">>> 开始处理回应...")
	var replyMsg common.ReplyMsg
	// 反序列化消息
	err := json.Unmarshal(payload,&replyMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		return
	}

	c.mutex.Lock()
	c.replyLog[replyMsg.NodeID] = &replyMsg
	rlen := len(c.replyLog)
	c.mutex.Unlock()
	if rlen >= c.countNeedReceiveMsgAmount(){
		fmt.Println(">>> 请求成功!!!")
	}
}

func handleStream(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			fmt.Println(">>> 客户端读取到数据...")
			replyChan <- str
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	for  {
		data, ok := <- dataChan
		if ok {
			_, err := rw.WriteString(fmt.Sprintf("%s\n", data))
			if err != nil {
				fmt.Println("Error writing to buffer")
				panic(err)
			}
			err = rw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}
		}
	}
}

func sendData(rw *bufio.ReadWriter) {
	for  {
		data, ok := <- sendDataChan
		if ok {
			_, err := rw.WriteString(fmt.Sprintf("%s\n", data))
			if err != nil {
				fmt.Println("Error writing to buffer")
				panic(err)
			}
			err = rw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}
		}
	}
}


func (c *Client) sendRequest(ctx context.Context, data string)  {
	fmt.Println(">>> 客户端准备request消息...")
	// 构建request
	req := common.Request{
		data,
		hex.EncodeToString(common.GenerateDigest(data)),
	}

	// 序列化pubKey
	marshalPubkey, err := crypto.MarshalPublicKey(c.keypair.pubkey)
	sendClient := common.SendClient{
		c.client.ID().Pretty(),
		marshalPubkey,
	}
	// 构建request消息
	reqMsg := common.RequestMsg{
		"solve",
		int(time.Now().Unix()),
		sendClient,
		req,
	}

	// 对发送的消息进行签名
	sig, err := c.signMessage(reqMsg)
	if err != nil{
		fmt.Printf("%v\n", err)
	}

	// 组合并发送消息
	c.send(ctx, common.ComposeMsg(common.HRequest, reqMsg, sig), c.findPrimaryNode())
	fmt.Println(">>> 客户端发送消息完成...")
}

func (c *Client) send(ctx context.Context, msg []byte, node KnownNode) {
	// 开始连接到主节点
	if err := c.client.Connect(ctx, node.h); err != nil{
		log.Println(">>> 连接到主节点失败")
	}

	// 打开stream
	s, err := c.client.NewStream(context.Background(), node.h.ID, protocol.ID(protocolID))
	if err != nil {
		fmt.Println(">>> 打开stream失败", err)
	}
	fmt.Println(">>> 开始连接到: ", node.h)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// 准备发送数据的通道
	go sendData(rw)

	// 序列化数据
	data, err := json.Marshal(msg)
	if err != nil{
		fmt.Println("序列化数据错误", err)
	}

	sendDataChan <- data
	close(sendDataChan)
}

func (c *Client) signMessage(msg interface{}) ([]byte, error){
	sig, err := common.SignMessage(msg, c.keypair.privkey)
	if err != nil{
		return nil, err
	}
	return sig, nil
}

// 确定主节点
func (c *Client) findPrimaryNode() KnownNode {
	fmt.Println(">>> 开始寻找主节点")

	if len(c.knownNodes) == 0{
		time.Sleep(time.Second)
	}
	nodeID := ViewID % len(c.knownNodes)
	fmt.Println(">>> 寻找主节点成功，主节点ID：", nodeID)
	return c.knownNodes[nodeID]
}


func (c *Client) countTolerateFaultNode() int {
	return (len(c.knownNodes) - 1) / 3
}

func (c *Client) countNeedReceiveMsgAmount() int {
	f := c.countTolerateFaultNode()
	return f+1
}


type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	// An hour might be a long long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(); err != nil {
		panic(err)
	}
	return n.PeerChan
}

func (c *Client) getAllKonwons(h host.Host) {
	rendezvousString := "meetme"
	peerChan := initMDNS(h, rendezvousString)
	for{
		v, ok := <- peerChan
		if ok{
			fmt.Println(">>> 发现节点: ", v)
			node := KnownNode{
				v,
			}
			c.knownNodes = append(c.knownNodes, node)
		}

	}
}