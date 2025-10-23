package file

import (
	"bufio"
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
