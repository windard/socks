package high_performance_networks

import (
	"crypto/md5"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/pbkdf2"

	"github.com/xtaci/kcp-go/v5"
)

func KCPEncryptedLocalServe() {
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

func RelayTCPToEncryptedKCP(client net.Conn) {
	// 高级的加密算法+socks5代理转发
	key := pbkdf2.Key(GlobalConfig.SecretKey, GlobalConfig.SecretSalt, 1024, 32, sha1.New)
	block, _ := kcp.NewSalsa20BlockCrypt(key)
	//block, _ := kcp.NewNoneBlockCrypt(nil)

	remote, err := kcp.DialWithOptions(GlobalConfig.RemoteAddr, block, 10, 3)
	if err != nil {
		log.Printf("dial with kcp error:" + err.Error())
		client.Close()
		return
	}

	log.Printf("receive from:[%s]%s relay to:[%s]%s",
		client.RemoteAddr().Network(), client.RemoteAddr().String(),
		remote.RemoteAddr().Network(), remote.RemoteAddr().String(),
	)
	Socks5Forward(client, remote)
}

func RelayEncryptedKCPToTCP(client net.Conn) {

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
	StreamForward(client, remote)
}

func KCPEncryptedRemoteServe() {
	// 高级的加密算法+socks5代理转发
	key := pbkdf2.Key(GlobalConfig.SecretKey, GlobalConfig.SecretSalt, 1024, 32, sha1.New)
	block, _ := kcp.NewSalsa20BlockCrypt(key)
	//block, _ := kcp.NewNoneBlockCrypt(nil)

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

		go Socks5Process(client)
	}
}

func KCPEncryptedServe() {
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
	GlobalConfig.SecretSalt = []byte(fmt.Sprintf("%x", sha1.Sum([]byte(*secretKey))))

	fmt.Printf("%s[%s] -> [%s]\n", *role, *listenAddr, *remoteAddr)
	if *role == "local" {
		KCPEncryptedLocalServe()
	} else {
		KCPEncryptedRemoteServe()
	}
}
