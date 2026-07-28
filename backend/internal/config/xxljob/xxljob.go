package xxljob

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/pwh19920920/butterfly/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xxl-job/xxl-job-executor-go"
)

const defaultServerAddr = "http://127.0.0.1:8080/xxl-job-admin"

// Config XXL-JOB 配置
type Config struct {
	ServerAddr  string `yaml:"serverAddr"`
	AccessToken string `yaml:"accessToken"`
}

type xxlJobConf struct {
	Xxl Config `yaml:"xxl"`
}

type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	logrus.Info(fmt.Sprintf("xxl-job - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	logrus.Error(fmt.Sprintf("xxl-job - "+format, a...))
}

// GetXxlJobExec 初始化并注册 XXL-JOB 执行器路由
func GetXxlJobExec() xxl.Executor {
	viper.SetDefault("xxl.serverAddr", defaultServerAddr)
	xxlConf := new(xxlJobConf)
	config.LoadConf(&xxlConf)

	ip, _ := GetLocalIP()
	serverAddr := server.GetConf().ServerAddr
	portIndex := strings.LastIndex(serverAddr, ":")
	executorPort := "8088"
	if portIndex >= 0 && portIndex+1 < len(serverAddr) {
		executorPort = serverAddr[portIndex+1:]
	}

	exec := xxl.NewExecutor(
		xxl.SetLogger(&logger{}),
		xxl.ServerAddr(xxlConf.Xxl.ServerAddr),
		xxl.ExecutorIp(ip),
		xxl.AccessToken(xxlConf.Xxl.AccessToken),
		xxl.ExecutorPort(executorPort),
		xxl.RegistryKey(server.GetConf().ServerName),
	)
	exec.Init()

	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "run", HandlerFunc: gin.WrapF(exec.RunTask)})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "kill", HandlerFunc: gin.WrapF(exec.KillTask)})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "log", HandlerFunc: gin.WrapF(exec.TaskLog)})
	server.RegisterRoute("", route)
	return exec
}

// GetLocalIP 获取本机非回环 IPv4
func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
			continue
		}
		addressList, err := item.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addressList {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("get network interface fail")
}
