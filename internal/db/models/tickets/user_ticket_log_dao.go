package tickets

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserTicketLogStateEnabled  = 1 // 已启用
	UserTicketLogStateDisabled = 0 // 已禁用
)

type UserTicketLogDAO dbs.DAO

func NewUserTicketLogDAO() *UserTicketLogDAO {
	return dbs.NewDAO(&UserTicketLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTicketLogs",
			Model:  new(UserTicketLog),
			PkName: "id",
		},
	}).(*UserTicketLogDAO)
}

var SharedUserTicketLogDAO *UserTicketLogDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketLogDAO = NewUserTicketLogDAO()
	})
}

// CreateLog 创建日志
func (this *UserTicketLogDAO) CreateLog(tx *dbs.Tx, adminId int64, userId int64, ticketId int64, status userconfigs.UserTicketStatus, comment string, isReadonly bool) (int64, error) {
	if ticketId <= 0 {
		return 0, errors.New("invalid ticketId")
	}

	var op = NewUserTicketLogOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.TicketId = ticketId
	op.Status = status
	op.Comment = comment
	op.IsReadonly = isReadonly
	op.State = UserTicketLogStateEnabled
	logId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = SharedUserTicketDAO.UpdateTicketStatus(tx, ticketId, status)
	if err != nil {
		return 0, err
	}
	return logId, nil
}

// CountTicketLogs 查询日志数量
func (this *UserTicketLogDAO) CountTicketLogs(tx *dbs.Tx, ticketId int64) (int64, error) {
	return this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		Count()
}

// ListTicketLogs 列出单页日志
func (this *UserTicketLogDAO) ListTicketLogs(tx *dbs.Tx, ticketId int64, offset int64, size int64) (result []*UserTicketLog, err error) {
	_, err = this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CheckUserLog 检查工单日志是否属于用户
func (this *UserTicketLogDAO) CheckUserLog(tx *dbs.Tx, userId int64, logId int64) error {
	if logId <= 0 {
		return models.ErrNotFound
	}
	b, err := this.Query(tx).
		Pk(logId).
		Attr("userId", userId).
		State(UserTicketStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return models.ErrNotFound
	}
	return nil
}

// CheckLogReadonly 检查日志是否为只读
func (this *UserTicketLogDAO) CheckLogReadonly(tx *dbs.Tx, logId int64) (bool, error) {
	return this.Query(tx).
		Result("isReadonly").
		Pk(logId).
		FindBoolCol()
}

// DisableLog 禁用日志
func (this *UserTicketLogDAO) DisableLog(tx *dbs.Tx, logId int64) error {
	return this.Query(tx).
		Pk(logId).
		Set("state", UserTicketLogStateDisabled).
		UpdateQuickly()
}

// FindLatestTicketLog 读取最新一条日志
func (this *UserTicketLogDAO) FindLatestTicketLog(tx *dbs.Tx, ticketId int64) (*UserTicketLog, error) {
	one, err := this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		DescPk().
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*UserTicketLog), nil
}
