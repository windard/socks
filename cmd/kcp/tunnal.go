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
	compatSocks5 bool
	kcpTunnelCmd = &cobra.Command{
		Use:   "tunnel",
		Short: "kcp tunnel to forward socket.",
		Long: `kcp tunnel to forward socket.
you can use compatSocks5 to compat with kcp socks5 connection
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

			if secretKey != "" {
				protocol.GlobalConfig.SecretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(secretKey))))
			}

			protocol.GlobalConfig.ListenAddr = listenAddr
			protocol.GlobalConfig.RemoteAddr = remoteAddr
			protocol.GlobalConfig.SecretSalt = []byte(fmt.Sprintf("%x", sha1.Sum([]byte(secretKey))))

			if role == "local" {
				fmt.Printf("[%s][tcp]Listen to:%s\n", role, listenAddr)
				if compatSocks5 {
					protocol.KCPLocalCompatSocks5Serve()
				} else {
					protocol.KCPLocalServe()
				}
			} else {
				fmt.Printf("[server][kcp]Listen on:%s\n", listenAddr)
				protocol.KCPRemoteServe()
			}

		},
	}
)

func init() {
	KcpCmd.AddCommand(kcpTunnelCmd)
	kcpTunnelCmd.Flags().StringVarP(&listenAddr, "localAddr", "l", "", "kcp proxy localAddr.")
	kcpTunnelCmd.Flags().StringVarP(&remoteAddr, "remoteAddr", "t", "", "kcp proxy remoteAddr.")
	kcpTunnelCmd.Flags().StringVarP(&role, "role", "r", "", "kcp proxy role.(local|server)")
	kcpTunnelCmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "kcp proxy secretKey.")
	kcpTunnelCmd.Flags().BoolVarP(&compatSocks5, "compatSocks5", "c", false, "kcp proxy compat with socks5.")
	kcpTunnelCmd.MarkFlagRequired("localAddr")
	kcpTunnelCmd.MarkFlagRequired("remoteAddr")
	kcpTunnelCmd.MarkFlagRequired("role")
	//kcpTunnelCmd.MarkFlagRequired("secretKey")
}
