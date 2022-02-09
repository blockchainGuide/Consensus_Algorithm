package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/spf13/cobra"
	"log"
)

func ServerCmd() *cobra.Command  {

	return serverCmd
}

var serverCmd = &cobra.Command{
	Use: "server",
	Short: "server manage",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Println("get param error: ", err)
		}
		// 创建server
		server := NewServer(port)
		// 开始server
		server.start()
	},
}

func init()  {
	serverCmd.Flags().IntP("port", "p", 0,"port")
}

type Server struct {
	node *Node
	url	string
}


func NewServer(port int) *Server {
	server := &Server{
		NewNode(port),
		"",
	}
	return server
}

func (s *Server) start()  {
	// 发现新的节点
	go s.getAllKonwons(s.node.node)
	s.node.Start()
	for  {

	}
}

// 服务端监听新新节点
func (s *Server) getAllKonwons(h host.Host) {
	rendezvousString := "meetme"
	peerChan := initMDNS(h, rendezvousString)
	for {
		v, ok := <- peerChan
		if ok{
			fmt.Println(">>> 发现节点: ", v)
			node := KnownNode{
				v,
			}
			if node.h.ID == s.node.node.ID(){
				continue
			}
			if node.h.ID.String() == s.node.clientNode.ID{
				fmt.Println("跳过客户端")
				continue
			}
			s.node.knownNodes = append(s.node.knownNodes, node)
		}
	}
}