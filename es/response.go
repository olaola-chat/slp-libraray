package es

type PutResponse struct {
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	Id      string `json:"_id"`
	Version int64  `json:"_version"`
	Result  string `json:"result"`
	Shards  shards `json:"_shards"`
	Created bool   `json:"created"`
}

type item struct {
	Index  string                 `json:"_index"`
	Type   string                 `json:"_type"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
	Sort   []float64              `json:"sort"`
}

type Total struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"`
}

type hits struct {
	Total int64 `json:"total"`
	//Total    Total   `json:"total"`
	MaxScore float64 `json:"max_score"`
	Hits     []item  `json:"hits"`
}

type shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

// Response 定义search返回数据结构
type SearchResponse struct {
	Took    int    `json:"took"`
	Timeout bool   `json:"timed_out"`
	Shared  shards `json:"_shards"`
	Hits    hits   `json:"hits"`
}

// Response 检索单个文档返回数据结构定义
type GetResponse struct {
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
	Docs []GetResponse `json:"docs"`
}
