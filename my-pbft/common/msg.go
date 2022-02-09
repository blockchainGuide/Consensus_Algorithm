package common

import (
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"reflect"
)


const headerLength = 12

type HeaderMsg string

const (
	HRequest 	HeaderMsg = "Request"
	HPrePrepare HeaderMsg = "PrePrepare"
	HPrepare	HeaderMsg = "Prepare"
	HCommit		HeaderMsg = "Commit"
	HReply		HeaderMsg = "Reply"
)

type Msg interface {
	String() string
}

type SendClient struct {
	ID string `json:"id"`
	PubKey []byte `json:"pub_key"`
}

type Request struct {
	// 消息内容
	Message		string	`json:"message"`
	// 消息摘要
	Digest		string	`json:"digest"`
}

func (msg Request) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

//<REQUEST, o, t, c>
type RequestMsg struct {
	// 请求的具体操作
	Operation		string	`json:"operation"`
	// 请求时客户端追加的时间戳
	Timestamp		int		`json:"timestamp"`
	// 客户端标识
	Client		SendClient		`json:"client"`
	// 包含消息内容以及消息摘要
	CRequest		Request	`json:"request"`
}

func (msg RequestMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

//<<PRE-PREPARE,v,n,d>,m>
type PrePrepareMsg struct {
	// 请求消息
	Request		RequestMsg	`json:"request"`
	// 签名
	Digest		string		`json:"digest"`
	// viewID
	ViewID     	int     	`json:"viewID"`
	// 队列ID
	SequenceID 	int     	`json:"sequenceID"`
}

func (msg PrePrepareMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

//<PREPARE, v, n, d, i>
type PrepareMsg struct {
	Digest		string		`json:"digest"`
	ViewID     	int       	`json:"viewID"`
	SequenceID 	int       	`json:"sequenceID"`
	NodeID		peer.ID			`json:"nodeid"`
}

func (msg PrepareMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

//<COMMIT, v, n, d, i>
type CommitMsg struct {
	Digest		string		`json:"digest"`
	ViewID     	int       	`json:"viewID"`
	SequenceID 	int       	`json:"sequenceID"`
	NodeID		peer.ID			`json:"nodeid"`
}

func (msg CommitMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

//<REPLY, v, t, c, i, r>
type ReplyMsg struct {
	ViewID     	int       	`json:"viewID"`
	Timestamp	int			`json:"timestamp"`
	ClientID	string		`json:"clientID"`
	NodeID		string			`json:"nodeid"`
	Result		string		`json:"result"`
}

func (msg ReplyMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg,"","	")
	return string(bmsg) + "\n"
}

// 组合消息
func ComposeMsg(header HeaderMsg, payload interface{}, sig []byte) []byte{
	var bpayload []byte
	var err error

	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		bpayload, err = json.Marshal(payload)
		if err != nil {
			panic(err)
		}
	case reflect.Slice:
		bpayload = payload.([]byte)
	default:
		panic(fmt.Errorf("not support type"))
	}

	b := make([]byte, headerLength)
	for i, h := range []byte(header){
		b[i] = h
	}
	res := make([]byte, headerLength+len(bpayload)+len(sig))
	copy(res[:headerLength],b)
	copy(res[headerLength:],bpayload)
	if len(sig) > 0 {
		copy(res[headerLength+len(bpayload):], sig)
	}
	return res
}

func SplitMsg(bmsg []byte) (HeaderMsg, []byte, []byte){
	// 消息头
	var header 		HeaderMsg
	// 消息载荷
	var payload		[]byte
	// 签名
	var signature 	[]byte
	hbyte := bmsg[:headerLength]
	hhbyte := make([]byte,0)
	for _, h := range hbyte{
		if h != byte(0){
			hhbyte = append(hhbyte,h)
		}
	}
	header = HeaderMsg(hhbyte)
	switch header {
	case HRequest, HPrePrepare, HPrepare, HCommit:
		payload = bmsg[headerLength:len(bmsg)-64]
		signature = bmsg[len(bmsg)-64:]
	case HReply:
		payload = bmsg[headerLength:]
		signature = []byte{}
	}
	return header, payload, signature
}
