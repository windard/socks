package kcp

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"os"

	protocol "github.com/windard/socks/protocol"

	"github.com/spf13/cobra"
)

var (
	kcpSocks5Cmd = &cobra.Command{
		Use:   "socks5",
		Short: "kcp socks5 implement socks5 proxy in kcp protocol.",
		Long: `kcp socks5 implement socks5 proxy in kcp protocol.
in local mode: only relay tcp to udp,
in server mode: accept udp, and implement socks5 protocol, so only localAddr take effect.
`,
		Run: func(cmd *cobra.Command, args []string) {

			if listenAddr == "" {
				fmt.Println("Please specify listen addr.")
				os.Exit(1)
			}

			if remoteAddr == "" {
				fmt.Println("Please specify remote addr.")
				os.Exit(1)
			}

			if secretKey == "" {
				fmt.Println("Please specify a secret.")
				os.Exit(1)
			}

			protocol.GlobalConfig.ListenAddr = listenAddr
			protocol.GlobalConfig.RemoteAddr = remoteAddr
			protocol.GlobalConfig.SecretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(secretKey))))
			protocol.GlobalConfig.SecretSalt = []byte(fmt.Sprintf("%x", sha1.Sum([]byte(secretKey))))

			if role == "local" {
				fmt.Printf("[%s][tcp]Connect to:%s\n", role, listenAddr)
				protocol.KCPEncryptedLocalServe()
			} else {
				fmt.Printf("[%s][kcp]Listen on:%s\n", role, listenAddr)
				protocol.KCPEncryptedRemoteServe()
			}

		},
	}
)

func init() {
	KcpCmd.AddCommand(kcpSocks5Cmd)
	kcpSocks5Cmd.Flags().StringVarP(&listenAddr, "localAddr", "l", "", "kcp proxy localAddr.")
	kcpSocks5Cmd.Flags().StringVarP(&remoteAddr, "remoteAddr", "t", "", "kcp proxy remoteAddr.")
	kcpSocks5Cmd.Flags().StringVarP(&role, "role", "r", "", "kcp proxy role.(local|server)")
	kcpSocks5Cmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "kcp proxy secretKey.")
	kcpSocks5Cmd.MarkFlagRequired("localAddr")
	kcpSocks5Cmd.MarkFlagRequired("remoteAddr")
	kcpSocks5Cmd.MarkFlagRequired("role")
	kcpSocks5Cmd.MarkFlagRequired("secretKey")
}
