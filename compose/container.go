package compose

import (
	"fmt"
	"os"
	"podman-compose/cli"
	"podman-compose/constant"
	"sync"
)

var ContainerList []cli.ListContainer
var lock sync.Mutex

func GetContainer(serviceName string) (cli.ListContainer, bool) {
	InitContainerList()

	for _, container := range ContainerList {
		v, ok := container.Labels[constant.LabelComposeServiceName]
		if ok && v == serviceName {
			return container, true
		}
	}
	return cli.ListContainer{}, false
}

/*
*
初始化容器列表
*/
func InitContainerList() {
	workDir, _ := os.Getwd()

	if ContainerList == nil {
		lock.Lock()
		if ContainerList == nil {
			containerListTmp := make([]cli.ListContainer, 0)
			all := true
			cs, err := cli.List(nil, &all, nil, nil, nil, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range cs {
				v, ok := c.Labels[constant.LabelComposeDir]
				if ok {
					if v == workDir {
						containerListTmp = append(containerListTmp, c)
					}
				}
			}
			ContainerList = containerListTmp
		}
		lock.Unlock()
	}
}
