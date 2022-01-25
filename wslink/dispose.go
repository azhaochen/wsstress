// Package server 压测启动
package wslink

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"
	"wsstress/model"
)

func TestOneWsConn(ctx context.Context, wsUrl string, c, n int) (err error) {
	// url提取服务，查询所有节点
	u, err := url.Parse(wsUrl)
	if err != nil {
		return err
	}
	// 查询服务节点
	nodeList := model.GetSvcInstances(u.Hostname())
	if len(nodeList) < 1 {
		return fmt.Errorf("ws polaris get %s instances failed.", u.Hostname())
	}
	wsUrl = fmt.Sprintf("ws://%s%s", nodeList[0], u.Path)
	fmt.Printf("ws start stress url:%s, c:%d, n:%d\n", wsUrl, c, n)

	var wgLink, wgStatistic sync.WaitGroup
	ch := make(chan *model.WsTrafficMsg, 10000) // 结束返回数据进行统计处理
	wgStatistic.Add(1)
	go func() {
		Statistic(ctx, ch, &wgStatistic, c, n)
	}()
	for i := 0; i < c; i++ {
		wgLink.Add(1)
		go func(connId int) {
			DealWithOneLink(ctx, wsUrl, connId, c, ch, &wgLink)
		}(i)
	}
	wgLink.Wait() // 全部协程结束
	// 延时1毫秒 确保数据都处理完成了
	time.Sleep(1 * time.Millisecond)
	close(ch)
	wgStatistic.Wait()
	fmt.Printf("ws finish\n")
	return nil
}

func DealWithOneLink(ctx context.Context, wsUrl string, linkID, totalID int, resultCh chan *model.WsTrafficMsg, wg *sync.WaitGroup) {
	defer wg.Done()
	ws := NewWebSocket(wsUrl)
	err := ws.GetConn()
	if err != nil {
		fmt.Printf("ws connect failed. err:%v\n", err)
		return
	}
	resultCh <- &model.WsTrafficMsg{ // 特殊消息，用于统计连接成功数目
		Tp:     model.WsConnectedTpFlag,
		ConnId: linkID,
	}
	// 读循环
	go func() {
		for {
			msg, err := ws.Read() // 阻塞读
			if err == io.EOF {    // 正常结束
				return
			}
			if linkID == 0 || linkID == totalID-1 {
				fmt.Printf("ws linkID:%d read msg:%s, err:%v\n", linkID, string(msg), err)
			}
			if err != nil && !strings.Contains(err.Error(), "closed") {
				fmt.Printf("ws read. err:%v\n", err)
				return
			}
			if len(msg) == 0 {
				return
			}
			responseJSON := &model.WsTrafficMsg{}
			err = json.Unmarshal(msg, responseJSON)
			if err != nil {
				fmt.Printf("ws read unmarshal. msg:%s, err:%v\n", string(msg), err)
				return
			}
			responseJSON.ConnId = linkID
			responseJSON.RecvTime = time.Now().Unix()
			resultCh <- responseJSON
		}
	}()
	// 退出：ctx超时退出，断线退出，异常返回退出
	for {
		select {
		case <-ctx.Done():
			_ = ws.Close() // 触发都循环退出
			return
		default:
		}
	}
}

// Statistic 统计协程
func Statistic(ctx context.Context, resultCh chan *model.WsTrafficMsg, wg *sync.WaitGroup, c, n int) {
	startTime := int64(0)
	lastTime := int64(0)
	successConn := 0
	resultMsg := make([]int64, c)
	defer func() {
		wg.Done()
		// 打印结果：
		fmt.Printf("======= ws stressws ======= c=%d\tn=%d\tsuccessConn=%d\ttotalConnTime=%d\n", c, n, successConn, lastTime-startTime)
		// 打印结果：接收消息耗时分布
		// 计算平均接收时间
		//var totalTime int64
		var msgCount int64
		var minTime, maxTime int64
		for _, v := range resultMsg {
			if v > 0 {
				if minTime == 0 {
					minTime = v
				}
				if minTime > v {
					minTime = v
				}
				if maxTime < v {
					maxTime = v
				}
				msgCount++
			}
		}
		fmt.Printf("======= ws stressws ======= MsgCount=%d\tminTime=%d\tmaxTime=%d\n", msgCount, minTime, maxTime)
	}()
	for {
		select {
		//case <-ctx.Done():
		//	return
		case msg, ok := <-resultCh:
			if !ok {
				return
			}
			if msg.Tp == model.WsConnectedTpFlag {
				if startTime == 0 {
					startTime = time.Now().Unix()
				}
				lastTime = time.Now().Unix()
				successConn++
			} else {
				//log.ErrorContextf(ctx, "ws channel recv msg: %+v", msg)
				resultMsg[msg.ConnId] = msg.RecvTime
			}
		default:
		}
	}
}
