// Package server 压测启动
package model

// WsTrafficMsg 返回数据结构体，返回值为json
type WsTrafficMsg struct {
	Tp       string `json:"tp"`
	HbtMsg   HbtMsg `json:"hbtMsg"`
	CtMsg    CtMsg  `json:"ctMsg"`
	ConnId   int    `json:"connId"`  // 本内部使用，标记是哪个
	RecvTime int64  `json:"revTime"` // 本内部使用，标记消息接收时间
}
type HbtMsg struct {
	Cmd     string   `json:"cmd"`
	Url     string   `json:"url"`
	Cookie  string   `json:"cookie"`
	AppName string   `json:"app_name"`
	Headers []string `json:"headers"`
	Extra   string   `json:"extra"`
}

const WsConnectedTpFlag = "connected" // 标记用，复用HbtMsg结构，cmd为这个值时，表示建立连接成功

type CtMsg struct {
	Cmd   string `json:"cmd"`
	Topic string `json:"topic"`
}

const (
	MaxRoutineTime = 10 * 60
)

// GetSvcInstances 获取polaris服务实例
func GetSvcInstances(svcname string) (list []string) {
	list = []string{"11.150.241.59:31272"}
	return list
}
