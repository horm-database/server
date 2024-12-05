package batch

const (
	RetBatchInsert        = 101 // 批量插入异常
	RetBatchDataUnMarshal = 102 // 批量结果解压缩失败
	RetHasBatchFailed     = 103 // 有异常待处理批量插入记录
	RetBatchFailedHandle  = 104 // 批量插入异常处理失败
	RetFormatDataError    = 105 // 格式化数据 error
)
