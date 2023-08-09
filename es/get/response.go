package get

// Response 检索单个文档返回数据结构定义
type Response struct {
	Index       string                 `json:"_index"`
	Type        string                 `json:"_type"`
	ID          string                 `json:"_id"`
	Version     int64                  `json:"_version"`
	SeqNo       int64                  `json:"_seq_no"`
	PrimaryTerm int64                  `json:"_primary_term"`
	Found       bool                   `json:"found"`
	Source      map[string]interface{} `json:"_source"`
}

// MResponse 检索多个文档返回数据结构定义
type MResponse struct {
	Docs []Response `json:"docs"`
}
