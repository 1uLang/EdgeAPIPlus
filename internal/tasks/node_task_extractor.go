package tasks

import (
	"github.com/1uLang/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNodeTaskExtractor(10 * time.Second).Start()
		})
	})
}

// NodeTaskExtractor 节点任务
type NodeTaskExtractor struct {
	BaseTask

	ticker *time.Ticker
}

func NewNodeTaskExtractor(duration time.Duration) *NodeTaskExtractor {
	return &NodeTaskExtractor{
		ticker: time.NewTicker(duration),
	}
}

func (this *NodeTaskExtractor) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("NodeTaskExtractor", err.Error())
		}
	}
}

func (this *NodeTaskExtractor) Loop() error {
	// 检查是否为主节点
	if !models.SharedAPINodeDAO.CheckAPINodeIsPrimaryWithoutErr() {
		return nil
	}

	// 这里不解锁，是为了让任务N秒钟之内只运行一次

	for _, role := range []string{nodeconfigs.NodeRoleNode, nodeconfigs.NodeRoleDNS} {
		err := models.SharedNodeTaskDAO.ExtractAllClusterTasks(nil, role)
		if err != nil {
			return err
		}
	}

	return nil
}
