package up

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"podman-compose/cli"
	"podman-compose/compose"
	"podman-compose/constant"
	"podman-compose/down"
	"podman-compose/registry"
	"podman-compose/util"
	"strconv"
	"strings"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start containers",
	Run:   up,
}

var detach = false

// 删除孤立项
var removeOrphans = false

func init() {
	upCmd.Flags().BoolVarP(&detach, "detach", "d", false, "daemon mode")
	upCmd.Flags().BoolVarP(&removeOrphans, "remove-orphans", "", false, "Remove containers for services not defined in the Compose file")
	registry.Commands = append(registry.Commands, upCmd)
}

func up(cmd *cobra.Command, args []string) {
	var err error
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	services := map[string]string{}
	for _, arg := range args {
		services[arg] = "1"
	}

	dockerCompose := compose.GetDockerCompose()

	//先统计服务总数
	//如果是 非 detach 模式， 则异步一起启动
	var serviceNum int
	var channel chan int
	if len(services) == 0 {
		serviceNum = len(dockerCompose.Services)
	} else {
		serviceNum = len(services)
	}
	channel = make(chan int, serviceNum)

	if len(services) == 0 {
		for serviceName := range dockerCompose.Services {
			if detach {
				serviceUp(serviceName, dockerCompose.Services[serviceName], channel)
			} else {
				go serviceUp(serviceName, dockerCompose.Services[serviceName], channel)
			}
		}
	} else {
		for service, _ := range services {
			serviceConfig, exist := dockerCompose.Services[service]
			if !exist {
				fmt.Printf("Service %s does not exist\n", service)
				channel <- 1
			} else {
				if detach {
					serviceUp(service, serviceConfig, channel)
				} else {
					go serviceUp(service, serviceConfig, channel)
				}
			}
		}
	}

	for {
		select {
		case <-channel:
			serviceNum--
			if serviceNum == 0 {
				fmt.Println("结束")
				break
			}
		}
	}
	//删除重复项
	down.RemoveOrphans(removeOrphans)

}

func serviceUp(serviceName string, service compose.ServiceConfig, channel chan int) {
	container, exist := compose.GetContainer(serviceName)
	upToDate := exist && isUpToDate(container, service)

	if upToDate {
		fmt.Println(compose.FormatServiceName(serviceName), " is up to date")
	} else {
		if exist {
			force := true
			fmt.Print(compose.FormatServiceName(serviceName) + " recreating... ")
			cli.Remove(container.ID, &force, nil)
		} else {
			fmt.Print(compose.FormatServiceName(serviceName) + " creating... ")
		}

		command, err := getCommand(serviceName, service)
		if err != nil {
			fmt.Println(serviceName, ":", err)
			os.Exit(1)
		}
		podmanCmd, err := exec.LookPath("podman")
		cmd := exec.Command(podmanCmd, command...)
		cmd.Stderr = os.Stdout
		err = cmd.Run()
		if err != nil {
			fmt.Println(serviceName, ":", err)
			return
		}
		fmt.Print(util.TextColor(32, "done"))
		fmt.Println()
	}
	channel <- 1
}

// 是否是最新
func isUpToDate(listContainer cli.ListContainer, service compose.ServiceConfig) bool {

	// 运行中 && 配置未修改 && 镜像也未修改  则是 up to date
	if listContainer.State == "running" {
		key := listContainer.Labels[constant.LabelConfigKey]
		expectKey := service.GetUnique()
		if key == expectKey {
			detail, err := cli.Inspect(listContainer.ID, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			image, err := cli.GetImage(service.Image, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if detail.Image == image.ID {
				return true
			}
		}
	}

	return false
}

/*
*
获取命令
*/
func getCommand(name string, service compose.ServiceConfig) ([]string, error) {
	command := []string{"run"}
	if detach {
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

	//标签
	command = append(command, "--label", constant.LabelComposeDir+"="+compose.GetComposeDir())
	command = append(command, "--label", constant.LabelComposeServiceName+"="+name)
	command = append(command, "--label", constant.LabelConfigKey+"="+service.GetUnique())

	//环境
	for k, v := range service.Environment {
		command = append(command, "--env", k+"="+v)
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
	_, err := cli.GetImage(image, nil)
	if err != nil {
		return command, err
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
