package tickets

// UserTicketLog 工单日志
type UserTicketLog struct {
	Id         uint64 `field:"id"`         // ID
	AdminId    uint64 `field:"adminId"`    // 管理员ID
	UserId     uint64 `field:"userId"`     // 用户ID
	TicketId   uint64 `field:"ticketId"`   // 工单ID
	Status     string `field:"status"`     // 状态
	Comment    string `field:"comment"`    // 回复内容
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	IsReadonly bool   `field:"isReadonly"` // 是否为只读
	State      uint8  `field:"state"`      // 状态
}

type UserTicketLogOperator struct {
	Id         interface{} // ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	TicketId   interface{} // 工单ID
	Status     interface{} // 状态
	Comment    interface{} // 回复内容
	CreatedAt  interface{} // 创建时间
	IsReadonly interface{} // 是否为只读
	State      interface{} // 状态
}

func NewUserTicketLogOperator() *UserTicketLogOperator {
	return &UserTicketLogOperator{}
}
