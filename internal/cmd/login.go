/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"honi/common"
	"honi/internal/auth"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login account",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.GetWechatBot().Login(); err != nil {
			slog.Error("登录失败", "err", err)
			return
		}

		_, err := http.Get(common.HostServer + common.RunBot)
		if err != nil {
			slog.Error("请求bot运行api失败", "err", err)
			return
		}
		slog.Info("登录成功")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
