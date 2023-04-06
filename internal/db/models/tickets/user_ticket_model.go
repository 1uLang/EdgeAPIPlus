package tickets

// UserTicket 工单
type UserTicket struct {
	Id         uint64 `field:"id"`         // ID
	CategoryId uint64 `field:"categoryId"` // 分类ID
	ToAdminId  uint64 `field:"toAdminId"`  // 指派的管理员ID
	UserId     uint64 `field:"userId"`     // 用户ID
	Subject    string `field:"subject"`    // 标题
	Body       string `field:"body"`       // 内容
	Status     string `field:"status"`     // 状态
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	LastLogAt  uint64 `field:"lastLogAt"`  // 最后日志时间
	State      uint8  `field:"state"`      // 状态
}

type UserTicketOperator struct {
	Id         interface{} // ID
	CategoryId interface{} // 分类ID
	ToAdminId  interface{} // 指派的管理员ID
	UserId     interface{} // 用户ID
	Subject    interface{} // 标题
	Body       interface{} // 内容
	Status     interface{} // 状态
	CreatedAt  interface{} // 创建时间
	LastLogAt  interface{} // 最后日志时间
	State      interface{} // 状态
}

func NewUserTicketOperator() *UserTicketOperator {
	return &UserTicketOperator{}
}
