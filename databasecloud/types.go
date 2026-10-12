package databasecloud

// Pager holds caller-owned query output; Records may be nil and metadata is not query input.
type Pager[T any] struct {
	// Records contains decoded records; check errors before using partial results.
	Records []*T `json:"records"`

	// Total is the reported matching count; count and data do not share an automatic snapshot.
	Total int64 `json:"total"`
	// Size echoes the requested page size.
	Size int `json:"size"`
	// Number echoes the requested one-based page number.
	Number int `json:"number"`
}
