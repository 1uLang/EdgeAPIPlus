package tickets

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

const (
	UserTicketStateEnabled  = 1 // 已启用
	UserTicketStateDisabled = 0 // 已禁用
)

type UserTicketDAO dbs.DAO

func NewUserTicketDAO() *UserTicketDAO {
	return dbs.NewDAO(&UserTicketDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTickets",
			Model:  new(UserTicket),
			PkName: "id",
		},
	}).(*UserTicketDAO)
}

var SharedUserTicketDAO *UserTicketDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketDAO = NewUserTicketDAO()
	})
}

// EnableUserTicket 启用条目
func (this *UserTicketDAO) EnableUserTicket(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketStateEnabled).
		Update()
	return err
}

// DisableUserTicket 禁用条目
func (this *UserTicketDAO) DisableUserTicket(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketStateDisabled).
		Update()
	return err
}

// FindEnabledUserTicket 查找启用中的条目
func (this *UserTicketDAO) FindEnabledUserTicket(tx *dbs.Tx, id int64) (*UserTicket, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserTicketStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTicket), err
}

// CreateTicket 创建工单
func (this *UserTicketDAO) CreateTicket(tx *dbs.Tx, userId int64, categoryId int64, subject string, body string) (int64, error) {
	var op = NewUserTicketOperator()
	op.UserId = userId
	op.CategoryId = categoryId
	op.Subject = subject
	op.Body = body
	op.Status = userconfigs.UserTicketStatusNone
	op.LastLogAt = time.Now().Unix()
	op.State = UserTicketStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateTicket 修改工单
func (this *UserTicketDAO) UpdateTicket(tx *dbs.Tx, ticketId int64, categoryId int64, subject string, body string) error {
	if ticketId <= 0 {
		return errors.New("invalid categoryId")
	}
	var op = NewUserTicketOperator()
	op.Id = ticketId
	op.CategoryId = categoryId
	op.Subject = subject
	op.Body = body
	return this.Save(tx, op)
}

// UpdateTicketStatus 设置工单状态
func (this *UserTicketDAO) UpdateTicketStatus(tx *dbs.Tx, ticketId int64, status userconfigs.UserTicketStatus) error {
	return this.Query(tx).
		Pk(ticketId).
		Set("status", status).
		Set("lastLogAt", time.Now().Unix()).
		UpdateQuickly()
}

// CountAllTickets 计算工单数量
func (this *UserTicketDAO) CountAllTickets(tx *dbs.Tx, userId int64, categoryId int64, status userconfigs.UserTicketStatus) (int64, error) {
	var query = this.Query(tx)
	query.State(UserTicketStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if categoryId > 0 {
		query.Attr("categoryId", categoryId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}
	return query.Count()
}

// ListTickets 列出单页工单
func (this *UserTicketDAO) ListTickets(tx *dbs.Tx, userId int64, categoryId int64, status userconfigs.UserTicketStatus, offset int64, size int64) (result []*UserTicket, err error) {
	var query = this.Query(tx)
	query.State(UserTicketStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if categoryId > 0 {
		query.Attr("categoryId", categoryId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// CheckUserTicket 检查工单是否属于用户
func (this *UserTicketDAO) CheckUserTicket(tx *dbs.Tx, userId int64, ticketId int64) error {
	if ticketId <= 0 {
		return models.ErrNotFound
	}
	b, err := this.Query(tx).
		Pk(ticketId).
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

// FindTicketStatus 查找工单状态
func (this *UserTicketDAO) FindTicketStatus(tx *dbs.Tx, ticketId int64) (string, error) {
	return this.Query(tx).
		Pk(ticketId).
		Result("status").
		FindStringCol("")
}
