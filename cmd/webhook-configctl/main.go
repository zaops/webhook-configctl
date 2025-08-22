package main

import (
	"fmt"
	"os"

	"webhook-configctl/internal/add"
	"webhook-configctl/internal/validate"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webhook-configctl",
	Short: "Webhook 配置管理工具",
	Long:  "用于管理 webhook.yaml 配置文件的 CLI 工具",
}

func init() {
	rootCmd.AddCommand(add.NewAddCommand())
	rootCmd.AddCommand(validate.NewValidateCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
