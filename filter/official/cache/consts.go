package cache

const ( // 缓存一致性方案
	ConsistencyTypeNone    = 0 //无需保障一致性（最终一致性，缓存到期就没了）
	ConsistencyTypeCount   = 1 //记录条数（针对的是只有新增类型的数据）
	ConsistencyTypeVersion = 2 //版本号（没有单条记录的并发问题）
	ConsistencyTypeQueue   = 3 //队列
	ConsistencyTypeLock    = 4 //锁
)

const ( // redis 缓存前缀
	PreFindCache = "data_" //数据缓存
)
