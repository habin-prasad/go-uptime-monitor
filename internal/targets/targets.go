package targets

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadFile(filePath string) ([]string, error) {
	var urls []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("open URL file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			urls = append(urls, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan URL file: %w", err)
		}

		return urls, nil
	}

	return nil, fmt.Errorf("provide a URL with -url or a file with -file")
}
