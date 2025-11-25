package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// ReadTextFile reads a .txt file and returns all lines as a slice of strings.
func ReadTextFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	fmt.Printf("[ReadTextFile] done to read file!\n")

	return lines, nil
}

// ReadJsonFile đọc file JSON và trả về map[string]any
func ReadJsonFile(filePath string) (map[string]any, error) {
	// Đọc toàn bộ nội dung file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadJsonArrayFile(filePath string) ([]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result []any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
