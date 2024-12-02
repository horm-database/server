package uniquekey

const ( // 唯一键自动生成类型
	UKAutoGenNo         = 0 //不自动生成
	UKAutoGenByDB       = 1 //存储引擎自增，比如 mysql 的 auto createment
	UKAutoGenByUStorage = 2 //由统一存储自动生成全局唯一的值（注意，如果需要统一存储生成，字段类型必须是字符长，长度必须>=32）
)
