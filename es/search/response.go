package search

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
type Response struct {
	Took    int    `json:"took"`
	Timeout bool   `json:"timed_out"`
	Shared  shards `json:"_shards"`
	Hits    hits   `json:"hits"`
}
