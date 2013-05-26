package goweb

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const configFile = "conf/app.conf"

type Config struct {
	keys map[string]string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		make(map[string]string),
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	inputread := bufio.NewReader(file)

	for {
		input, _, err := inputread.ReadLine()
		if err == io.EOF {
			break
		}
		line := string(input)
		keyval := strings.SplitN(line, "=", 2)
		if len(keyval) == 2 {
			cfg.keys[strings.TrimSpace(keyval[0])] = strings.TrimSpace(keyval[1])
		}
	}

	return cfg, nil
}

func (c *Config) GetValue(key string) string {
	return c.keys[key]
}
