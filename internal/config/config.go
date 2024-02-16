package config

type Config struct {
	Api           Api    `json:"api"`
	LogLevel      string `json:"log_level"`
	IsDevelopment bool   `env:"DEVELOPMENT"`
}

type Api struct {
	Host         string `json:"host"`
	TCPPort      string `json:"tcp_port"`
	HTTPPort     string `json:"http_port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}
