package databasecloud

type Pager[T any] struct {
	Records []*T  `json:"records"` // 响应数据
	Total   int64 `json:"total"`   // 响应总记录数

	Size int `json:"size"` // 请求每页记录数
	Num  int `json:"num"`  // 请求页码 从1开始
}
