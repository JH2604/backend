package model

// LostItem 是"数据库实体"（Entity）：它和数据库表 lost_items 一一对应。
// 它的标签分两类：
//   gorm:"..."  → 给数据库看的（列名、主键、长度）
//   json:"..."  → 给前端看的（JSON 字段名）
//   binding:"..." → 给 Gin 的校验器看的（请求参数必须满足什么条件）
type LostItem struct {
	// ID：主键。gorm:"primaryKey" 告诉 GORM 这是主键、自增（类比数组的唯一下标）
	// json:"id" 表示前端传/收的字段名叫 id
	ID uint `gorm:"primaryKey" json:"id"`

	// Title：标题
	// binding:"required,notblank" —— 逗号分隔的多个校验，【必须全部满足】才算通过：
	//    required → 不能是"零值"：字符串里就是不能是 ""（而且字段不能缺）
	//    notblank → 去掉两端的空白字符后，必须还有内容
	//    为什么非加 notblank 不可？因为 " "（一个空格）长度是 1，
	//    在 required 眼里【不是空值】，能大摇大摆溜进去 → 脏数据（就是队友测到的那个 bug）
	Title string `json:"title" binding:"required,notblank"`

	// Location：地点（可选，前端不传就是空字符串）
	Location string `json:"location"`

	// Desc：详细描述（可选）
	Desc string `json:"desc"`
}
