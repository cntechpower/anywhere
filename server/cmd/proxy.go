package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/server/rpc/handler"

	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "proxy admin interface",
	Long:  `proxy admin interface.`,
}

//args for add proxy config command
var addProxyWhiteListIps string
var addProxyIsWhiteListOn bool

var proxyAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add proxy config",
	Long:  `add a proxy config.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.AddProxyConfig(userName, groupName, remotePort, localAddr, addProxyIsWhiteListOn, addProxyWhiteListIps); err != nil {
			fmt.Printf("error adding proxy config : %v\n", err)
		}
	},
}

//args for del proxy config command
var groupName, localAddr string
var remotePort int

var proxyDelCmd = &cobra.Command{
	Use:   "del",
	Short: "delete proxy config",
	Long:  `delete a proxy config.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.RemoveProxyConfig(userName, groupName, remotePort, localAddr); err != nil {
			fmt.Printf("error deleting proxy config : %v\n", err)
		}
	},
}

var proxyListCmd = &cobra.Command{
	Use:   "list",
	Short: "list proxy configs",
	Long:  `add a proxy config.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.ListProxyConfigs(); err != nil {
			fmt.Printf("error query proxy config list: %v\n", err)
		}
	},
}

var proxyLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load proxy configs",
	Long:  `load proxy configs from config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.LoadProxyConfigFile(); err != nil {
			fmt.Printf("error load proxy config: %v\n", err)
		}
	},
}

var proxySaveCmd = &cobra.Command{
	Use:   "save",
	Short: "save proxy configs",
	Long:  `save proxy configs to config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.SaveProxyConfigToFile(); err != nil {
			fmt.Printf("error save proxy config: %v\n", err)
		}
	},
}

//args for add proxy config command
var updateProxyAgentId, updateProxyLocalAddr, updateProxyWhiteListIps string
var updateProxyIsWhiteListOn bool

var proxyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update proxy config white list",
	Long:  `update proxy config white list.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.UpdateProxyConfigWhiteList(userName, updateProxyAgentId, updateProxyLocalAddr,
			updateProxyWhiteListIps, updateProxyIsWhiteListOn); err != nil {
			fmt.Printf("error save proxy config: %v\n", err)
		}
	},
}

func GetProxyCmd() *cobra.Command {
	proxyCmd.PersistentFlags().StringVar(&userName, "user", "", "user name")
	proxyCmd.PersistentFlags().StringVar(&groupName, "group", "", "group name")
	proxyAddCmd.PersistentFlags().IntVar(&remotePort, "remote-addr", 0, "remote port")
	proxyAddCmd.PersistentFlags().StringVar(&localAddr, "local-addr", "127.0.0.1:80", "local addr")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyWhiteListIps, "white-list", "", "local port")
	proxyAddCmd.PersistentFlags().BoolVar(&addProxyIsWhiteListOn, "enable-wl", false, "enable white list or not")
	proxyDelCmd.PersistentFlags().IntVar(&remotePort, "remote-port", 0, "del from which remotePort")
	proxyDelCmd.PersistentFlags().StringVar(&localAddr, "local-addr", "", "del from which localAddr")
	proxyUpdateCmd.PersistentFlags().StringVar(&updateProxyLocalAddr, "local-addr", "127.0.0.1:80", "local addr")
	proxyUpdateCmd.PersistentFlags().StringVar(&updateProxyWhiteListIps, "white-list", "", "local port")
	proxyUpdateCmd.PersistentFlags().BoolVar(&updateProxyIsWhiteListOn, "enable-wl", false, "enable white list or not")
	proxyCmd.AddCommand(proxyListCmd)
	proxyCmd.AddCommand(proxyAddCmd)
	proxyCmd.AddCommand(proxyDelCmd)
	proxyCmd.AddCommand(proxyLoadCmd)
	proxyCmd.AddCommand(proxySaveCmd)
	proxyCmd.AddCommand(proxyUpdateCmd)
	return proxyCmd
}
