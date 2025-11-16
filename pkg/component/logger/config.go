package logger

type LoggerConfig struct {
	Level   string `config:"level"`
	Format  string `config:"format"`
	Output  string `config:"output"`
	NoColor bool   `config:"no_color"`
}
