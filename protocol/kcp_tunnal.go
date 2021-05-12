package high_performance_networks

import (
	"crypto/md5"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/xtaci/kcp-go/v5"
)

type KCPConfig struct {
	ListenAddr string
	RemoteAddr string
	SecretKey  []byte
	SecretSalt []byte
}

var GlobalConfig = KCPConfig{}

func KCPLocalServe() {
	server, err := net.Listen("tcp", GlobalConfig.ListenAddr)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}

		go RelayTCPToKCP(client)
	}
}

func KCPLocalCompatSocks5Serve() {
	server, err := net.Listen("tcp", GlobalConfig.ListenAddr)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}

		go RelayTCPToEncryptedKCP(client)
	}
}

func RelayTCPToKCP(client net.Conn) {
	// 仅做了简单的流式加密
	block, _ := kcp.NewNoneBlockCrypt(nil)
	sess, err := kcp.DialWithOptions(GlobalConfig.RemoteAddr, block, 10, 3)
	if err != nil {
		log.Printf("dial with kcp error:" + err.Error())
		client.Close()
		return
	}

	remote, err := NewChacha20Stream(GlobalConfig.SecretKey, sess)
	if err != nil {
		log.Printf("new stream error:" + err.Error())
		client.Close()
		return
	}

	log.Printf("receive from:[%s]%s relay to:[%s]%s",
		client.RemoteAddr().Network(), client.RemoteAddr().String(),
		sess.RemoteAddr().Network(), sess.RemoteAddr().String(),
	)
	StreamForward(client, remote)
}

func RelayKCPToTCP(client net.Conn) {
	// 仅做了简单的流式解密
	src, err := NewChacha20Stream(GlobalConfig.SecretKey, client)
	if err != nil {
		log.Printf("new stream error:" + err.Error())
		client.Close()
		return
	}

	remote, err := net.Dial("tcp", GlobalConfig.RemoteAddr)
	if err != nil {
		log.Printf("dial with kcp error:" + err.Error())
		client.Close()
		return
	}

	log.Printf("receive from:[%s]%s relay to:[%s]%s",
		client.RemoteAddr().Network(), client.RemoteAddr().String(),
		remote.RemoteAddr().Network(), remote.RemoteAddr().String(),
	)
	StreamForward(src, remote)
}

func KCPRemoteServe() {
	block, _ := kcp.NewNoneBlockCrypt(nil)
	listener, err := kcp.ListenWithOptions(GlobalConfig.ListenAddr, block, 10, 3)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := listener.AcceptKCP()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}

		go RelayKCPToTCP(client)
	}
}

func KCPServe() {
	listenAddr := flag.String("listenAddr", "127.0.0.1:9800", "listen addr")
	remoteAddr := flag.String("remoteAddr", "127.0.0.1:9801", "remote addr")
	role := flag.String("role", "local", "serve role: local or remote")
	secretKey := flag.String("secret", "", "secret key")
	flag.Parse()

	if *secretKey == "" {
		fmt.Printf("Please sepcify a secret.\n")
		return
	}

	GlobalConfig.RemoteAddr = *remoteAddr
	GlobalConfig.ListenAddr = *listenAddr
	GlobalConfig.SecretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(*secretKey))))

	fmt.Printf("%s[%s] -> [%s]\n", *role, *listenAddr, *remoteAddr)
	if *role == "local" {
		KCPLocalServe()
	} else {
		KCPRemoteServe()
	}
}
