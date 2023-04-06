package tickets

// UserTicketCategory 工单分类
type UserTicketCategory struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 分类名
	IsOn  bool   `field:"isOn"`  // 是否启用
	Order uint32 `field:"order"` // 排序
	State uint8  `field:"state"` // 状态
}

type UserTicketCategoryOperator struct {
	Id    interface{} // ID
	Name  interface{} // 分类名
	IsOn  interface{} // 是否启用
	Order interface{} // 排序
	State interface{} // 状态
}

func NewUserTicketCategoryOperator() *UserTicketCategoryOperator {
	return &UserTicketCategoryOperator{}
}
