package compose

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func (config *ServiceConfig) GetUnique() string {

	key1 := config.toBinary()
	key2 := key1 + "-key2"
	m := md5.New()
	m.Write([]byte(key1))
	md51 := hex.EncodeToString(m.Sum(nil))

	m = md5.New()
	m.Write([]byte(key2))
	md52 := hex.EncodeToString(m.Sum(nil))
	return md51 + "-" + md52
}

func (config *ServiceConfig) toBinary() string {
	var rs []string
	rs = append(rs, config.Image, config.Restart, config.Entrypoint,
		config.WorkingDir,
		strconv.FormatFloat(config.Deploy.Limits.CPUs, 'E', -1, 64),
		config.Deploy.Limits.Memory,
		config.ContainerName)
	rs = append(rs, config.Command...)
	rs = append(rs, config.Ports...)

	var envKeys []string
	for k, _ := range config.Environment {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)
	for _, key := range envKeys {
		rs = append(rs, key, config.Environment[key])
	}

	rs = append(rs, config.Volumes...)
	return strings.Join(rs, "-")
}
