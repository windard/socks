package cmd

import (
	"fmt"
	"os"

	"github.com/windard/socks/cmd/kcp"

	"github.com/spf13/cobra"
)

const (
	Version = "v0.1.3"
)

var (
	RootCmd = &cobra.Command{
		Use:     "socks",
		Version: Version,
		Short:   "Socks Tool implement socks protocol, such as socks4 and socks5, also provide kcp encryption extension",
		Long: `Socks Tool

Socks Tool implement socks protocol, 
such as socks4 and socks5, 
also provide kcp encryption extension.`,
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(kcp.KcpCmd)
}

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of socks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
