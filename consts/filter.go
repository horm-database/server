package consts

const ( // 插件类型
	PreFilter   = 1 // 前置插件
	PostFilter  = 2 // 后置插件
	DeferFilter = 3 // 延迟插件（query 最后执行，不改变返回结果，而且每个 defer 都会被执行）
)

const ( // 插件配置类型
	FilterConfigTypeBool        = 1  // bool
	FilterConfigTypeString      = 2  // string
	FilterConfigTypeInt         = 3  // int
	FilterConfigTypeUint        = 4  // uint
	FilterConfigTypeFloat       = 5  // float
	FilterConfigTypeBytes       = 6  // bytes
	FilterConfigTypeEnum        = 7  // 枚举（单选）
	FilterConfigTypeMultiChoice = 8  // 多选
	FilterConfigTypeTime        = 9  // 时间
	FilterConfigTypeArray       = 10 // 数组
	FilterConfigTypeMap         = 11 // map
	FilterConfigTypeMultiConf   = 12 // 配置数组
)

const ( // 插件动作类型
	ActionTypeExec = 1 // 1-执行插件
	ActionTypeSkip = 2 // 2-跳过插件
)

const ( // 规则类型
	CondTypeAny = 1 // 1-任一规则(条件)
	CondTypeAll = 2 // 2-所有规则（条件）
)
