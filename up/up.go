package up

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"podman-compose/compose"
	"podman-compose/registry"
	"strconv"
	"strings"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start containers",
	Run:   up,
}

var daemon bool = false

func init() {
	upCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "daemon mode")
	registry.Commands = append(registry.Commands, upCmd)
}

func up(cmd *cobra.Command, args []string) {
	dockerCompose := compose.GetDockerCompose()
	for serviceName := range dockerCompose.Services {
		fmt.Println("服务名称: ", serviceName)
		command, err := getCommand(serviceName, dockerCompose.Services[serviceName])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
		podmanCmd, err := exec.LookPath("podman")
		cmd := exec.Command(podmanCmd, command...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
	}
}

/*
*
获取命令
*/
func getCommand(name string, service compose.ServiceConfig) ([]string, error) {
	command := []string{"run"}
	if daemon {
		command = append(command, "--detach")
	}
	//端口
	command, err := formatPort(command, &service)
	if err != nil {
		return command, err
	}

	//挂载卷
	command, err = formatVolumes(command, &service)
	if err != nil {
		return command, err
	}

	//容器名称
	fmt.Println("service.ContainerName ", service.ContainerName)
	if strings.TrimSpace(service.ContainerName) != "" {
		command = append(command, "--name", service.ContainerName)
	}

	//workdir
	if strings.TrimSpace(service.WorkingDir) != "" {
		command = append(command, "--workdir", service.WorkingDir)
	}

	//Entrypoint
	if strings.TrimSpace(service.Entrypoint) != "" {
		command = append(command, "--entrypoint", service.Entrypoint)
	}

	// restart
	if strings.TrimSpace(service.Restart) != "" {
		command = append(command, "--restart", service.Restart)
	}

	//镜像
	command, err = formatImage(command, &service)
	if err != nil {
		return command, err
	}

	//容器命令
	command, err = formatContainerCommand(command, &service)
	if err != nil {
		return command, err
	}
	return command, nil
}

// 镜像
func formatContainerCommand(command []string, service *compose.ServiceConfig) ([]string, error) {
	if len(service.Command) > 0 {
		for _, cmd := range service.Command {
			command = append(command, cmd)
		}
	}
	return command, nil
}

// 镜像
func formatImage(command []string, service *compose.ServiceConfig) ([]string, error) {
	image := strings.TrimSpace(service.Image)
	if image == "" {
		return command, errors.New("image is required")
	}
	command = append(command, service.Image)
	return command, nil
}

// 挂载卷
func formatVolumes(command []string, service *compose.ServiceConfig) ([]string, error) {
	for _, volume := range service.Volumes {
		command = append(command, "-v", volume)
	}
	return command, nil
}

// 组装端口
func formatPort(command []string, service *compose.ServiceConfig) ([]string, error) {
	for _, port := range service.Ports {
		pair := strings.Split(port, ":")
		if len(pair) == 1 {
			_, err := strconv.Atoi(pair[0])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
		}
		if len(pair) == 2 {
			_, err := strconv.Atoi(pair[0])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
			_, err = strconv.Atoi(pair[1])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
		}
		command = append(command, "-p", port)
	}

	return command, nil
}
