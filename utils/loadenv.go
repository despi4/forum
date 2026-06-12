package utils

import (
	"os"
	"strings"
)

func Loadenv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes := make([]byte, 1024)

	n, err := file.Read(bytes)
	if err != nil {
		return err
	}

	env_data := strings.Split(string(bytes[:n]), "\n")

	for _, str := range env_data {
		if len(str) == 0 || strings.HasPrefix(str, "#") {
			continue
		}

		parts := strings.SplitN(str, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		os.Setenv(key, value)
	}

	return nil
}
