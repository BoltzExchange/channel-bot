package main

func main() {
	cfg := loadConfig()
	initLogger(cfg.LogFile)
	logConfig(cfg)
}
