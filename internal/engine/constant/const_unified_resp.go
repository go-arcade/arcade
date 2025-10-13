package constant

// UnifiedResponse 统一响应
const (
	// DETAIL Detail 用于设置响应数据，例如查询，分页等，需要返回数据
	// e.g: c.Set(DETAIL, value)
	DETAIL = "detail"

	// OPERATION Operation 用于设置响应数据，例如新增，修改，删除等，不需要返回数据，只返回操作结果
	// e.g: c.Set(OPERATION, "")
	OPERATION = "operation"
)
