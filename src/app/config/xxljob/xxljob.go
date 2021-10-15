package xxljob

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/config"
	"github.com/pwh19920920/butterfly/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xxl-job/xxl-job-executor-go"
	"net"
	"strings"
)

const defaultServerAddr = "http://127.0.0.1:8080/xxl-job-admin"

type Config struct {
	ServerAddr  string `yaml:"serverAddr"`
	AccessToken string `yaml:"accessToken"`
}

var xxlConf *xxlJobConf

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

func GetXxlJobExec() xxl.Executor {
	// 默认配置
	viper.SetDefault("xxl.serverAddr", defaultServerAddr)

	// 加载配置
	xxlConf = new(xxlJobConf)
	config.LoadConf(&xxlConf, config.GetOptions().ConfigFilePath)

	// 获取本地ip
	ip, _ := GetLocalIP()

	// 取出端口
	severAddr := server.GetConf().ServerAddr
	portIndex := strings.LastIndex(severAddr, ":")
	executorPort := severAddr[portIndex+1:]

	//初始化执行器
	exec := xxl.NewExecutor(
		xxl.SetLogger(&logger{}),
		xxl.ServerAddr(xxlConf.Xxl.ServerAddr),
		xxl.ExecutorIp(ip),
		xxl.AccessToken(xxlConf.Xxl.AccessToken),     //请求令牌(默认为空)
		xxl.ExecutorPort(executorPort),               //默认9999（此处要与gin服务启动port必需一至）
		xxl.RegistryKey(server.GetConf().ServerName), //执行器名称
	)
	exec.Init()

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "run", HandlerFunc: gin.WrapF(exec.RunTask)})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "kill", HandlerFunc: gin.WrapF(exec.KillTask)})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "log", HandlerFunc: gin.WrapF(exec.TaskLog)})
	server.RegisterRoute("", route)

	// 返回执行器
	return exec
}

func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if item.Flags&net.FlagLoopback != 0 {
			continue // loop back interface
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
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("get network interface fail")
}
