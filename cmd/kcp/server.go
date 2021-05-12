package kcp

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"os"

	protocol "github.com/windard/socks/protocol"

	"github.com/spf13/cobra"
)

const defaultAddr = "127.0.0.1:1080"

var (
	listenAddr string
	remoteAddr string
	role       string
	secretKey  string

	KcpCmd = &cobra.Command{
		Use:   "kcp",
		Short: "kcp protocol implement echo server.",
		Long: `kcp protocol implement echo server.
see https://github.com/skywind3000/kcp.
no output if in wrong protocol or wrong secretKey.
`,
		Run: func(cmd *cobra.Command, args []string) {

			if listenAddr == "" {
				listenAddr = defaultAddr
			}

			if secretKey == "" {
				fmt.Println("Please specify a secret.")
				os.Exit(1)
			}

			protocol.GlobalConfig.ListenAddr = listenAddr
			protocol.GlobalConfig.SecretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(secretKey))))
			protocol.GlobalConfig.SecretSalt = []byte(fmt.Sprintf("%x", sha1.Sum([]byte(secretKey))))

			if role == "local" {
				fmt.Printf("[%s][kcp]Connect to:%s\n", role, listenAddr)
				protocol.KCPClient()
			} else {
				fmt.Printf("[%s][kcp]Listen on:%s\n", role, listenAddr)
				protocol.KCPServer()
			}

		},
	}
)

func init() {
	KcpCmd.Flags().StringVarP(&listenAddr, "listenAddr", "l", "", "kcp proxy listenAddr.")
	KcpCmd.Flags().StringVarP(&role, "role", "r", "", "kcp proxy role.(local|server)")
	KcpCmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "kcp proxy secretKey.(required)")
	KcpCmd.MarkFlagRequired("role")
	KcpCmd.MarkFlagRequired("secretKey")
}
