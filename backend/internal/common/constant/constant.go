package constant

var ContextUser = "X-AUTH-USER"

// MaxAggregateRows 聚合采集单次查询返回的最大行数。
// MySQL/MongoDB/URL 三路聚合取数统一使用此上限，超出截断并记录警告，防止海量结果 OOM。
const MaxAggregateRows = 10000
