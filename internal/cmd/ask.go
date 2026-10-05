/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"honi/internal/agent"
	"strings"

	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "单次对话",
	RunE: func(cmd *cobra.Command, args []string) error {
		message := strings.Join(args, ",")
		answer, err := agent.GetAgent().Ask(message)
		if err != nil {
			return err
		}

		fmt.Println(answer)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
