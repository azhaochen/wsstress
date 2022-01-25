// Package server 压测启动
package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"
	"wsstress/model"
	"wsstress/wslink"
)

var (
	concurrency  uint64 = 1       // 并发数
	totalNumber  uint64 = 1       // 请求数(单个并发/协程)
	totalTimeout uint64 = 1       // 压测时间
	svcNodeNum   uint64 = 1       // 压测目前节点数目
	debugStr     string = "false" // 是否是debug
	service      string = ""      // 压测的url 目前支持 ws://, wss://
)

func init() {
	flag.Uint64Var(&concurrency, "c", concurrency, "并发数：2")
	flag.Uint64Var(&totalNumber, "n", totalNumber, "请求数(单个并发/协程): 1")
	flag.Uint64Var(&totalTimeout, "t", totalTimeout, "压测最大时间s: 20")
	flag.Uint64Var(&svcNodeNum, "node", svcNodeNum, "压测目前节点数目: 1")
	flag.StringVar(&debugStr, "d", debugStr, "调试模式: true")
	flag.StringVar(&service, "svc", service, "压测服务: ws://11.150.241.59:31272/NoLogin")
	// 解析参数
	flag.Parse()
}

//go:generate go build main.go
func main() {
	runtime.GOMAXPROCS(1)
	if concurrency == 0 || totalNumber == 0 || service == "" || totalTimeout < 5 {
		flag.Usage()
		return
	}
	if totalTimeout > model.MaxRoutineTime {
		totalTimeout = model.MaxRoutineTime / 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(totalTimeout)*time.Second)
	defer cancel()
	err := wslink.TestOneWsConn(ctx, service, int(concurrency), int(totalNumber))
	fmt.Printf("main done. %+v", err)
}
