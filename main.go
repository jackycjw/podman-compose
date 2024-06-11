package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"podman-compose/compose"
	"podman-compose/registry"
	_ "podman-compose/up"
)

var rootCmd = &cobra.Command{
	Use: "podman-compose",

	Short: `Define and run multi-container applications with Docker.

Usage:
  docker-compose [-f <arg>...] [--profile <name>...] [options] [--] [COMMAND] [ARGS...]
  docker-compose -h|--help`,
}

func main() {
	for _, cmd := range registry.Commands {
		rootCmd.AddCommand(cmd)
	}
	//初始化Compose文件
	err := compose.InitCompose()
	if err != nil {
		fmt.Println(err)
	}
	if err = rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
