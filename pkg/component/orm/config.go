package orm

type ORMConfig struct {
	Driver   string `config:"driver"`
	Host     string `config:"host"`
	User     string `config:"user"`
	Password string `config:"password"`
	Name     string `config:"name"`
	Port     int    `config:"port"`
	SSLMode  string `config:"sslmode"`
}
