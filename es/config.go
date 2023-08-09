package es

// Config 定义了ES服务器访问的配置
type Config struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}
