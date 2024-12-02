package batch

const ( // 批量插入开关
	FlushOpen  = 1 //开
	FlushClose = 0 //关
)

const (
	InsertPopMax = 70000
)

const ( // 读写优先
	ReadPriority  = 0
	WritePriority = 1
)

const BufferRedis = "buffer" //BufferRedis 批量插入缓冲区

const ( // redis 缓存前缀
	PreBatchInsertBuff     = "BatchBuff"     //批量插入缓冲区
	PreBatchExchange       = "BatchEx"       //交换缓冲区
	PreBatchMutex          = "BatchMutex"    //缓冲区处理互斥
	PreBatchInsertFailBuff = "BatchFailBuff" //批量插入失败缓冲区
)
