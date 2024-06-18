package startup

import (
	"fmt"
	"os"
	"os/exec"
	"podman-compose/cli"
	"podman-compose/util"
	"time"
)

var retry = 0

func StartUp() {
	all := true
	containerList, err := cli.List(nil, &all, nil, nil, nil, nil)
	if err != nil {
		if retry < 20 {
			time.Sleep(time.Second * 1)
			retry++
			StartUp()
		} else {
			return
		}
	}

	for _, container := range containerList {
		if container.Exited {
			detail, err := cli.Inspect(container.ID, nil)
			if err != nil {
				fmt.Println(err)
			}
			if detail.HostConfig.RestartPolicy.Name == "always" {
				fmt.Print(container.Names[0] + "[" + container.ID + "] starting... ")
				//启动
				cli.Start(container.ID, nil)

				//重新加载网络
				{
					podmanCmd, err := exec.LookPath("podman")
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					cmd := exec.Command(podmanCmd, "network", "reload", container.ID)
					cmd.Stderr = os.Stdout
					err = cmd.Run()
				}
				fmt.Print(util.TextColor(32, "done"))
				fmt.Println()
			}
		}
	}
}

func getReloadCommand(container cli.ListContainer) []string {
	return []string{"network", "reload", container.ID}
}
