package es

type errorCause struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type errorDetail struct {
	Reason    string       `json:"reason"`
	Type      string       `json:"type"`
	RootCause []errorCause `json:"root_cause"`
}

// ErrorResponse 定义ES返回错误的数据结构
type ErrorResponse struct {
	Status int64       `json:"status"`
	Error  errorDetail `json:"error"`
}
