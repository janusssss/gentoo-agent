/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"honi/config"
	"log/slog"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	var list bool
	conf := config.GetConfig()
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置程序运行参数",
	}
	cmd.Flags().StringVar(&conf.Provider, "provider", conf.Provider, "设置LLM厂商或兼容厂商")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "查看配置列表")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		conf := config.GetConfig()
		if list || cmd.Flags().NFlag() == 0 {
			buff, err := toml.Marshal(conf)
			if err != nil {
				return err
			}
			fmt.Print(string(buff))
			return nil
		}

		if conf.Compare() {
			slog.Warn("config is duplicate")
			return nil
		}
		if err := conf.WriteConfig(); err != nil {
			return err
		}

		slog.Info("config write done")
		return nil
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newConfigCmd())
}
