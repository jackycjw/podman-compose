package main

import (
	"encoding/binary"
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

type Test struct {
	Name  string
	Kv    map[string]string
	Label []string
}

func main() {
	test := Test{Name: "陈家文", Kv: map[string]string{"a": "b", "c": "d"}, Label: []string{"e", "h", "j"}}

	binary.Write(os.Stdout, binary.LittleEndian, test)

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
