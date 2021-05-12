package protocol

import (
	"crypto/md5"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/xtaci/kcp-go/v5"
)

func KCPServer() {
	key := pbkdf2.Key(GlobalConfig.SecretKey, GlobalConfig.SecretSalt, 1024, 32, sha1.New)
	block, _ := kcp.NewSalsa20BlockCrypt(key)
	//block, _ := kcp.NewNoneBlockCrypt(nil)

	listener, err := kcp.ListenWithOptions(GlobalConfig.ListenAddr, block, 10, 3)
	if err != nil {
		log.Printf("Listen failed:%v\n", err)
		return
	}

	log.Printf("Ready for connecting.\n")
	for {
		client, err := listener.AcceptKCP()
		if err != nil {
			log.Printf("Accept error:" + err.Error())
			continue
		}

		remoteAddr := client.RemoteAddr().String()
		remoteNetwork := client.RemoteAddr().Network()
		log.Printf("Connection from: [%s]%s\n", remoteNetwork, remoteAddr)
		go HandleEcho(client)
	}
}

func HandleEcho(client net.Conn) {
	buf := make([]byte, 2048)

	for {
		n, err := client.Read(buf)
		if err != nil {
			log.Printf("Read error:" + err.Error())
			return
		}

		log.Printf("Receive Message:%s", buf[:n])
		_, err = client.Write(buf[:n])
		if err != nil {
			log.Printf("Write error:" + err.Error())
			return
		}
	}
}

func KCPClient() {
	key := pbkdf2.Key(GlobalConfig.SecretKey, GlobalConfig.SecretSalt, 1024, 32, sha1.New)
	block, _ := kcp.NewSalsa20BlockCrypt(key)
	//block, _ := kcp.NewNoneBlockCrypt(nil)

	sess, err := kcp.DialWithOptions(GlobalConfig.ListenAddr, block, 10, 3)
	if err != nil {
		log.Printf("Dial error:%+v\n", err)
		return
	}

	for i := 0; i < 10; i++ {
		data := time.Now().String()
		log.Printf("【send】:%s", data)
		_, err := sess.Write([]byte(data))
		if err != nil {
			log.Printf("Write error:" + err.Error())
			return
		}

		buf := make([]byte, 2048)
		// Read will block until Buffer full or Newline(\n)
		n, err := sess.Read(buf)
		if err != nil {
			log.Printf("Read error:" + err.Error())
			return
		}
		log.Printf("【recv】:%s", buf[:n])
		time.Sleep(time.Second)
	}
}

func KCPDemo() {
	listenAddr := flag.String("listenAddr", "127.0.0.1:8976", "listen addr")
	role := flag.String("role", "server", "serve role: server or client")
	secretKey := flag.String("secret", "", "secret key")
	flag.Parse()

	if *secretKey == "" {
		fmt.Printf("Please sepcify a secret.\n")
		return
	}

	GlobalConfig.ListenAddr = *listenAddr
	GlobalConfig.SecretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(*secretKey))))
	GlobalConfig.SecretSalt = []byte(fmt.Sprintf("%x", sha1.Sum([]byte(*secretKey))))

	if *role == "client" {
		fmt.Printf("[%s]Connect to:%s\n", *role, *listenAddr)
		KCPClient()
	} else {
		fmt.Printf("[%s]Listen on:%s\n", *role, *listenAddr)
		KCPServer()
	}
}
