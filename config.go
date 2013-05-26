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

func (c *Config) GetString(key string, defaultvalue string) string {
	return ToString(c.keys[key], defaultvalue)
}

func (c *Config) GetInt(key string, defaultvalue int) int {
	return ToInt(c.keys[key], defaultvalue)
}

func (c *Config) GetBool(key string, defaultvalue bool) bool {
	return ToBool(c.keys[key], defaultvalue)
}

func (c *Config) GetMap() map[string]string {
	return c.keys
}
