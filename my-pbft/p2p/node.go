package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
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
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"my-pbft/common"
	"sync"
	"time"
)

var dataChan = make(chan []byte)
var receiveChan = make(chan string)
var broadcastChan = make(chan []byte)
var replyChan = make(chan []byte)

const ViewID = 0
var protocolID string = "/chat/1.1.0"
type Node struct {
	knownNodes		[]KnownNode
	clientNode		ClientNode
	// 队列ID
	sequenceID 		int
	node 			host.Host
	View			int
	// 消息队列
	msgQueue		chan []byte
	// 密码对
	keypair			Keypair
	// 消息日志
	msgLog			*MsgLog
	// 请求池
	requestPool		map[string]*common.RequestMsg
	// 互斥锁
	mutex			sync.Mutex
}

type ClientNode struct {
	ID string
	pubKey crypto.PubKey
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

type MsgLog struct {
	preprepareLog	map[string]map[string]bool
	prepareLog		map[string]map[string]bool
	commitLog		map[string]map[string]bool
	replyLog	map[string]bool
}

func NewNode(port int) *Node {
	// 生成密钥对
	r := rand.Reader
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, r)
	if err != nil{
		log.Println(err)
	}
	pubKey := prvKey.GetPublic()
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", "127.0.0.1", port))

	// 创建libp2p节点
	h, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Println(err)
	}

	h.SetStreamHandler(protocol.ID(protocolID), handleStream)
	fmt.Printf(">>> 创建客户端p2p节点成功，客户端多路地址是: /ip4/%s/tcp/%v/p2p/%s\n", "0.0.0.0", port, h.ID().Pretty())

	keyPair := Keypair{
		privkey: prvKey,
		pubkey: pubKey,
	}

	// 创建node
	return &Node{
		[]KnownNode{},
		ClientNode{},
		0,
		h,
		ViewID,
		make(chan []byte),
		keyPair,
		&MsgLog{
			make(map[string]map[string]bool),
			make(map[string]map[string]bool),
			make(map[string]map[string]bool),
			make(map[string]bool),
		},
		make(map[string]*common.RequestMsg),
		sync.Mutex{},
	}
}

func (node *Node)Start()  {
	// 处理消息
	go node.handleMsg()
}

func (node *Node) handleMsg() {
	fmt.Println(">>> 启动节点，等待接收消息...")
	for  {

		//  待改进 todo
		rawData := <- receiveChan

		//msg := <- node.msgQueue // todo
		//node.msgQueue <- data // todo

		// 反序列得到的数据
		var data []byte
		err := json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			fmt.Println("反序列消息化失败：", err)
			return
		}

		// 分割消息，分别处理不同的消息
		header, payload, sign:= common.SplitMsg(data)
		switch header {
		case common.HRequest:
			node.handleRequest(payload, sign)
		case common.HPrePrepare:
			node.handlePrePrepare(payload, sign)
		case common.HPrepare:
			node.handlePrepare(payload, sign)
		case common.HCommit:
			node.handleCommit(payload, sign)
		default:
			fmt.Println("===============无法处理对应的消息============")
		}
	}
}

func (node *Node) handleRequest(payload []byte, sig []byte)  {
	fmt.Println(">>> 主节点接收request消息...")
	var request common.RequestMsg
	var prePrepareMsg common.PrePrepareMsg

	// 反序列化请求消息
	err := json.Unmarshal(payload, &request)
	if err != nil{
		log.Println("反序列化request错误: ", err)
		return
	}

	// 设置节点的客户端
	clientPubKey, err := crypto.UnmarshalPublicKey(request.Client.PubKey)
	if err != nil{
		fmt.Println(">>> 反序列化客户端公钥失败", err)
	}
	clientNode := ClientNode{
		request.Client.ID,
		clientPubKey,
	}
	node.clientNode = clientNode


	// 校验request的摘要
	vdig := common.VerifyDigest(request.CRequest.Message, request.CRequest.Digest)
	if vdig == false {
		fmt.Printf("验证摘要错误\n")
		return
	}

	// 校验request的签名
	_, err  = common.VerifySignatrue(request, sig, clientPubKey)
	if err != nil  {
		fmt.Printf("验证签名错误:%v\n", err)
		return
	}

	// 添加进请求池
	node.mutex.Lock()
	node.requestPool[request.CRequest.Digest] = &request
	seqID := node.getSequenceID()
	node.mutex.Unlock()

	// 构建pre-Prepare消息
	prePrepareMsg = common.PrePrepareMsg{
		request,
		request.CRequest.Digest,
		ViewID,
		seqID,
	}

	// 消息签名
	msgSig, err:= node.signMessage(prePrepareMsg)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}

	// 消息组合
	msg := common.ComposeMsg(common.HPrePrepare, prePrepareMsg, msgSig)

	// 日志处理
	node.mutex.Lock()
	if node.msgLog.preprepareLog[prePrepareMsg.Digest] == nil {
		node.msgLog.preprepareLog[prePrepareMsg.Digest] = make(map[string]bool)
	}
	node.msgLog.preprepareLog[prePrepareMsg.Digest][node.node.ID().String()] = true
	node.mutex.Unlock()

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil{
		fmt.Println("序列化request消息出错", err)
		return
	}

	fmt.Println(">>> 主节点广播prePrepare消息...")
	// 广播消息
	node.broadcast(data)
}

func (node *Node) handlePrePrepare(payload []byte, sig []byte) {
	fmt.Println(">>> 副节点开始接收prePrepare消息...")
	// 反序列化prePrepare消息
	var prePrepareMsg common.PrePrepareMsg
	err := json.Unmarshal(payload,&prePrepareMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		return
	}

	// 找到主节点的公钥
	pnodeId := node.findPrimaryNode()
	pubKey, err := pnodeId.h.ID.ExtractPublicKey()
	if err != nil {
		fmt.Println("获取主节点的公钥失败", err)
		return
	}

	// 校验消息签名
	_, err = common.VerifySignatrue(prePrepareMsg, sig, pubKey)
	if err != nil  {
		fmt.Printf("验证主节点签名错误:%v\n", err)
		return
	}

	// 校验消息的摘要
	if prePrepareMsg.Digest != prePrepareMsg.Request.CRequest.Digest {
		fmt.Printf("校验摘要错误\n")
		return
	}

	node.mutex.Lock()
	node.requestPool[prePrepareMsg.Request.CRequest.Digest] = &prePrepareMsg.Request
	node.mutex.Unlock()

	// 校验request的摘要
	err = node.verifyRequestDigest(prePrepareMsg.Digest)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}

	node.mutex.Lock()
	node.requestPool[prePrepareMsg.Request.CRequest.Digest] = &prePrepareMsg.Request
	node.mutex.Unlock()

	err = node.verifyRequestDigest(prePrepareMsg.Digest)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}

	node.mutex.Lock()
	if node.msgLog.preprepareLog[prePrepareMsg.Digest] == nil {
		node.msgLog.preprepareLog[prePrepareMsg.Digest] = make(map[string]bool)
	}
	node.msgLog.preprepareLog[prePrepareMsg.Digest][node.node.ID().String()] = true
	node.mutex.Unlock()

	// 构建prePare消息
	prepareMsg := common.PrepareMsg{
		prePrepareMsg.Digest,
		ViewID,
		prePrepareMsg.SequenceID,
		node.node.ID(),
	}

	// 签名
	msgSig, err := common.SignMessage(prepareMsg, node.keypair.privkey)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}
	// 消息组合
	sendMsg := common.ComposeMsg(common.HPrepare,prepareMsg,msgSig)

	// 序列化消息
	data, err := json.Marshal(sendMsg)
	if err != nil{
		fmt.Println("序列化prepare消息出错", err)
		return
	}

	fmt.Println(">>> 副节点广播prepare消息...")
	node.broadcast(data)
}

func (node *Node) handlePrepare(payload []byte, sig []byte) {
	fmt.Println(">>> 副节点开始接收prepare消息...")

	// 反序列化prepare消息
	var prepareMsg common.PrepareMsg
	err := json.Unmarshal(payload,&prepareMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		return
	}

	// 得到节点的公钥
	pnodeID := prepareMsg.NodeID
	pubKey, err:= findNodePubkey(pnodeID)
	if err != nil {
		fmt.Println("获取主节点的公钥失败", err)
		return
	}

	_, err = common.VerifySignatrue(prepareMsg, sig, pubKey)
	if err != nil  {
		fmt.Printf("校验签名prepare消息错误:%v\n", err)
		return
	}

	err = node.verifyRequestDigest(prepareMsg.Digest)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}


	// todo
	// verify prepareMsg's digest is equal to preprepareMsg's digest
	//exist := node.msgLog.preprepareLog[prepareMsg.Digest][pnodeID.String()]
	//fmt.Println("是否存在========================", exist)
	//fmt.Println(prepareMsg.Digest)
	//
	//if !exist {
	//	fmt.Printf("this digest's preprepare msg by %s not existed\n", pnodeID)
	//	return
	//}

	// 日记记录
	node.mutex.Lock()
	if node.msgLog.prepareLog[prepareMsg.Digest] == nil {
		node.msgLog.prepareLog[prepareMsg.Digest] = make(map[string]bool)
	}
	node.msgLog.prepareLog[prepareMsg.Digest][prepareMsg.NodeID.String()] = true
	node.mutex.Unlock()

	// if receive prepare msg >= 2f +1, then broadcast commit msg
	limit := node.countNeedReceiveMsgAmount()
	sum, err  := node.findVerifiedPrepareMsgCount(prepareMsg.Digest)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		return
	}
	if sum >= limit {
		//send commit msg
		commitMsg := common.CommitMsg{
			prepareMsg.Digest,
			prepareMsg.ViewID,
			prepareMsg.SequenceID,
			node.node.ID(),
		}
		sig, err := node.signMessage(commitMsg)
		if err != nil{
			fmt.Printf("sign message happened error:%v\n", err)
		}
		sendMsg := common.ComposeMsg(common.HCommit,commitMsg,sig)

		data, err := json.Marshal(sendMsg)
		if err != nil{
			fmt.Println("序列化commit消息出错", err)
		}

		node.broadcast(data)
		fmt.Println(">>> 副节点广播commit消息成功")
	}
}

func (node *Node) handleCommit(payload []byte, sig []byte) {
	fmt.Println(">>> 副节点开始接收commit消息")

	// 反序列化消息
	var commitMsg common.CommitMsg
	err := json.Unmarshal(payload,&commitMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
	}

	msgPubKey, err := findNodePubkey(commitMsg.NodeID)
	if err != nil{
		fmt.Println(err)
		return
	}

	verify, err := common.VerifySignatrue(commitMsg, sig, msgPubKey)
	if err != nil  {
		fmt.Printf("verify signature failed:%v\n", err)
		return
	}
	if verify == false {
		fmt.Printf("verify signature failed\n")
		return
	}

	err = node.verifyRequestDigest(commitMsg.Digest)
	if err != nil{
		fmt.Printf("%v\n", err)
		return
	}

	node.mutex.Lock()
	if node.msgLog.commitLog[commitMsg.Digest] == nil {
		node.msgLog.commitLog[commitMsg.Digest] = make(map[string]bool)
	}
	node.msgLog.commitLog[commitMsg.Digest][commitMsg.NodeID.String()] = true
	node.mutex.Unlock()
	// if receive commit msg >= 2f +1, then send reply msg to client
	limit := node.countNeedReceiveMsgAmount()
	sum, err := node.findVerifiedCommitMsgCount(commitMsg.Digest)
	if err != nil{
		fmt.Printf("error happened:%v", err)
		return
	}

	if sum >= limit {
		// if already send reply msg, then do nothing
		node.mutex.Lock()
		exist := node.msgLog.replyLog[commitMsg.Digest]
		node.mutex.Unlock()
		if exist == true {
			return
		}
		// send reply msg
		node.mutex.Lock()
		requestMsg := node.requestPool[commitMsg.Digest]
		node.mutex.Unlock()
		fmt.Printf("operstion:%s  message:%s executed... \n",requestMsg.Operation, requestMsg.CRequest.Message)
		done := fmt.Sprintf("operstion:%s  message:%s done ",requestMsg.Operation, requestMsg.CRequest.Message)
		replyMsg := common.ReplyMsg{
			node.View,
			int(time.Now().Unix()),
			requestMsg.Client.ID,
			node.node.ID().String(),
			done,
		}

		hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", requestMsg.Client.ID))
		clientNode, err := peer.AddrInfoFromString(hostAddr.String())
		fmt.Println(">>> 客户端地址：", clientNode.ID.Pretty())
		if err != nil{
			fmt.Println(err)
			return
		}

		fmt.Println(">>> 开始向客户端回复数据...")
		sendMsg := common.ComposeMsg(common.HReply,replyMsg,[]byte{})
		data, err := json.Marshal(sendMsg)
		if err != nil{
			fmt.Println("序列化commit消息出错", err)
		}
		node.reply(context.Background(), data, *clientNode)
		node.mutex.Lock()
		node.msgLog.replyLog[commitMsg.Digest] = true
		node.mutex.Unlock()
		fmt.Println(">>> 回复客户端成功...")
	}
}


func (node *Node) findVerifiedPrepareMsgCount(digest string) (int, error){
	sum:=0
	node.mutex.Lock()
	for _, exist := range node.msgLog.prepareLog[digest]{
		if exist == true {
			sum++
		}
	}
	node.mutex.Unlock()
	return sum, nil
}

func (node *Node) countTolerateFaultNode() int {
	return (len(node.knownNodes) - 1) / 3
}

func (node *Node) countNeedReceiveMsgAmount() int {
	f := node.countTolerateFaultNode()
	return 2*f+1
}

func (node *Node) verifyRequestDigest(digest string) error {
	node.mutex.Lock()
	_, ok := node.requestPool[digest]
	if !ok {
		node.mutex.Unlock()
		return 	fmt.Errorf("verify request digest failed\n")

	}
	node.mutex.Unlock()
	return nil
}

func (c *Node) findPrimaryNode() KnownNode {

	if len(c.knownNodes) == 0{
		time.Sleep(time.Second)
	}
	nodeID := ViewID % len(c.knownNodes)
	return c.knownNodes[nodeID]
}

func (node *Node) broadcast (data []byte)  {
	for _, knownNode := range node.knownNodes{
		if knownNode.h.ID != node.node.ID(){
			if knownNode.h.ID.Pretty() != node.clientNode.ID{
				// todo 待优化
				fmt.Println(">>> 准备发送的节点ID", knownNode.h)
				node.send(context.Background(), data, knownNode.h)
			}
		}
	}
}

func (node *Node) send(ctx context.Context, msg []byte, h peer.AddrInfo) {
	// 连接到主节点
	if err := node.node.Connect(ctx, h); err != nil{
		log.Println(">>> 连接到节点失败", err)
		return
	}
	fmt.Println(">>> 连接节点成功，节点ID", h.ID)

	s, err := node.node.NewStream(context.Background(), h.ID, protocol.ID(protocolID))
	if err != nil {
		fmt.Println(">>> 创建流失败", err)
		return
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go broadcastData(rw)

	broadcastChan <- msg
}

func (node *Node) reply(ctx context.Context, msg []byte, h peer.AddrInfo) {
	// 连接到主节点
	if err := node.node.Connect(ctx, h); err != nil{
		log.Println(">>> 连接到节点失败", err)
		return
	}
	fmt.Println(">>> 连接节点成功，节点ID", h.ID)

	s, err := node.node.NewStream(context.Background(), h.ID, protocol.ID(protocolID))
	if err != nil {
		fmt.Println(">>> 创建流失败", err)
		return
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go replyData(rw)

	replyChan <- msg
}


func (node *Node) getSequenceID() int{
	seq := node.sequenceID
	node.sequenceID++
	return seq
}

func (node *Node) signMessage(msg interface{}) ([]byte, error){
	sig, err := common.SignMessage(msg, node.keypair.privkey)
	if err != nil{
		return nil, err
	}
	return sig, nil
}

func (node *Node) findVerifiedCommitMsgCount(digest string) (int, error){
	sum:=0
	node.mutex.Lock()
	for _, exist := range node.msgLog.commitLog[digest]{

		if exist == true{
			sum++
		}
	}
	node.mutex.Unlock()
	return sum, nil
}

func findNodePubkey(id peer.ID) (crypto.PubKey, error) {
	return id.ExtractPublicKey()
}

func handleStream(stream network.Stream) {
	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
		}

		if str == "" {
			return
		}
		if str != "\n" {
			receiveChan <- str
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	for  {
		data, ok := <- dataChan
		if ok{
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

func broadcastData(rw *bufio.ReadWriter) {
	for  {
		data, ok := <- broadcastChan
		if ok{
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

func replyData(rw *bufio.ReadWriter) {
	for  {
		data, ok := <- replyChan
		if ok{
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

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

//Initialize the MDNS service
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



