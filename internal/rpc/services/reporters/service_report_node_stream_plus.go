//go:build plus
// +build plus

package reporters

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// CommandRequest 命令请求相关
type CommandRequest struct {
	Id          int64
	Code        string
	CommandJSON []byte
}

type CommandRequestWaiting struct {
	Timestamp int64
	Chan      chan *pb.ReportNodeStreamMessage
}

func (this *CommandRequestWaiting) Close() {
	defer func() {
		recover()
	}()

	close(this.Chan)
}

var responseChanMap = map[int64]*CommandRequestWaiting{} // request id => response
var commandRequestId = int64(0)

var nodeLocker = &sync.Mutex{}
var requestChanMap = map[int64]chan *CommandRequest{} // node id => chan

func NextCommandRequestId() int64 {
	return atomic.AddInt64(&commandRequestId, 1)
}
func init() {
	dbs.OnReadyDone(func() {
		// 清理WaitingChannelMap
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			for range ticker.C {
				nodeLocker.Lock()
				for requestId, request := range responseChanMap {
					if time.Now().Unix()-request.Timestamp > 3600 {
						responseChanMap[requestId].Close()
						delete(responseChanMap, requestId)
					}
				}
				nodeLocker.Unlock()
			}
		}()

		// 自动同步连接到本API节点的Report节点任务
		go func() {
			defer func() {
				_ = recover()
			}()

			// TODO 未来支持同步边缘节点
			var ticker = time.NewTicker(3 * time.Second)
			for range ticker.C {
				nodeIds, err := models.SharedNodeTaskDAO.FindAllDoingNodeIds(nil, nodeconfigs.NodeRoleReport)
				if err != nil {
					remotelogs.Error("ReportNodeService_SYNC", err.Error())
					continue
				}
				nodeLocker.Lock()
				for _, nodeId := range nodeIds {
					c, ok := requestChanMap[nodeId]
					if ok {
						select {
						case c <- &CommandRequest{
							Id:          NextCommandRequestId(),
							Code:        reporterconfigs.MessageCodeNewNodeTask,
							CommandJSON: nil,
						}:
						default:

						}
					}
				}
				nodeLocker.Unlock()
			}
		}()
	})
}

// ReportNodeStream 节点stream
func (this *ReportNodeService) ReportNodeStream(server pb.ReportNodeService_ReportNodeStreamServer) error {
	// TODO 使用此stream快速通知Reporter节点更新
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(server.Context(), rpcutils.UserTypeReport)
	if err != nil {
		return err
	}

	tx := this.NullTx()
	err = validateClient(tx, nodeId, server.Context())
	if err != nil {
		return err
	}

	// 返回连接成功
	{
		apiConfig, err := configs.SharedAPIConfig()
		if err != nil {
			return err
		}
		connectedMessage := &reporterconfigs.ConnectedAPINodeMessage{APINodeId: apiConfig.NumberId()}
		connectedMessageJSON, err := json.Marshal(connectedMessage)
		if err != nil {
			return errors.Wrap(err)
		}
		err = server.Send(&pb.ReportNodeStreamMessage{
			Code:     reporterconfigs.MessageCodeConnectedAPINode,
			DataJSON: connectedMessageJSON,
		})
		if err != nil {
			return err
		}
	}

	//logs.Println("[RPC]accepted ns node '" + types.String(nodeId) + "' connection")

	// 标记为活跃状态
	oldIsActive, err := models.SharedReportNodeDAO.FindNodeActive(tx, nodeId)
	if err != nil {
		return err
	}
	if !oldIsActive {
		err = models.SharedReportNodeDAO.UpdateNodeActive(tx, nodeId, true)
		if err != nil {
			return err
		}

		// 发送恢复消息
		nodeName, err := models.SharedReportNodeDAO.FindReportNodeName(tx, nodeId)
		if err != nil {
			return err
		}
		subject := "区域监控节点\"" + nodeName + "\"已经恢复在线"
		msg := "区域监控节点\"" + nodeName + "\"已经恢复在线"
		err = models.SharedMessageDAO.CreateNodeMessage(tx, nodeconfigs.NodeRoleReport, 0, nodeId, models.MessageTypeReportNodeActive, models.MessageLevelSuccess, subject, msg, nil, false)
		if err != nil {
			return err
		}
	}

	nodeLocker.Lock()
	requestChan, ok := requestChanMap[nodeId]
	if !ok {
		requestChan = make(chan *CommandRequest, 1024)
		requestChanMap[nodeId] = requestChan
	}
	nodeLocker.Unlock()

	defer func() {
		nodeLocker.Lock()
		delete(requestChanMap, nodeId)
		nodeLocker.Unlock()
	}()

	// 发送请求
	go func() {
		for {
			select {
			case <-server.Context().Done():
				return
			case commandRequest := <-requestChan:
				// logs.Println("[RPC]sending command '" + commandRequest.Code + "' to node '" + strconv.FormatInt(nodeId, 10) + "'")
				retries := 3 // 错误重试次数
				for i := 0; i < retries; i++ {
					err := server.Send(&pb.ReportNodeStreamMessage{
						RequestId: commandRequest.Id,
						Code:      commandRequest.Code,
						DataJSON:  commandRequest.CommandJSON,
					})
					if err != nil {
						if i == retries-1 {
							logs.Println("[RPC]send command '" + commandRequest.Code + "' failed: " + err.Error())
						} else {
							time.Sleep(1 * time.Second)
						}
					} else {
						break
					}
				}
			}
		}
	}()

	// 接受请求
	for {
		req, err := server.Recv()
		if err != nil {
			// 修改节点状态
			err1 := models.SharedReportNodeDAO.UpdateNodeActive(tx, nodeId, false)
			if err1 != nil {
				logs.Println(err1.Error())
			}

			return err
		}

		func(req *pb.ReportNodeStreamMessage) {
			// 因为 responseChan.Chan 有被关闭的风险，所以我们使用recover防止panic
			defer func() {
				recover()
			}()

			nodeLocker.Lock()
			responseChan, ok := responseChanMap[req.RequestId]
			if ok {
				select {
				case responseChan.Chan <- req:
				default:

				}
			}
			nodeLocker.Unlock()
		}(req)
	}
}

// SendCommandToReportNode 向节点发送命令
func (this *ReportNodeService) SendCommandToReportNode(ctx context.Context, req *pb.ReportNodeStreamMessage) (*pb.ReportNodeStreamMessage, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	nodeId := req.ReportNodeId
	if nodeId <= 0 {
		return nil, errors.New("node id should not be less than 0")
	}

	nodeLocker.Lock()
	requestChan, ok := requestChanMap[nodeId]
	nodeLocker.Unlock()

	if !ok {
		return &pb.ReportNodeStreamMessage{
			RequestId: req.RequestId,
			IsOk:      false,
			Message:   "node '" + strconv.FormatInt(nodeId, 10) + "' not connected yet",
		}, nil
	}

	req.RequestId = NextCommandRequestId()

	select {
	case requestChan <- &CommandRequest{
		Id:          req.RequestId,
		Code:        req.Code,
		CommandJSON: req.DataJSON,
	}:
		// 加入到等待队列中
		respChan := make(chan *pb.ReportNodeStreamMessage, 1)
		waiting := &CommandRequestWaiting{
			Timestamp: time.Now().Unix(),
			Chan:      respChan,
		}

		nodeLocker.Lock()
		responseChanMap[req.RequestId] = waiting
		nodeLocker.Unlock()

		// 等待响应
		timeoutSeconds := req.TimeoutSeconds
		if timeoutSeconds <= 0 {
			timeoutSeconds = 10
		}
		timeout := time.NewTimer(time.Duration(timeoutSeconds) * time.Second)
		select {
		case resp := <-respChan:
			// 从队列中删除
			nodeLocker.Lock()
			delete(responseChanMap, req.RequestId)
			waiting.Close()
			nodeLocker.Unlock()

			if resp == nil {
				return &pb.ReportNodeStreamMessage{
					RequestId: req.RequestId,
					Code:      req.Code,
					Message:   "response timeout",
					IsOk:      false,
				}, nil
			}

			return resp, nil
		case <-timeout.C:
			// 从队列中删除
			nodeLocker.Lock()
			delete(responseChanMap, req.RequestId)
			waiting.Close()
			nodeLocker.Unlock()

			return &pb.ReportNodeStreamMessage{
				RequestId: req.RequestId,
				Code:      req.Code,
				Message:   "response timeout over " + fmt.Sprintf("%d", timeoutSeconds) + " seconds",
				IsOk:      false,
			}, nil
		}
	default:
		return &pb.ReportNodeStreamMessage{
			RequestId: req.RequestId,
			Code:      req.Code,
			Message:   "command queue is full over " + strconv.Itoa(len(requestChan)),
			IsOk:      false,
		}, nil
	}
}
