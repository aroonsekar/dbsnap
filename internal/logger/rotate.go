package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Rotate truncates log lines older than the specified day limit (e.g., 30 days).
func Rotate(backupPath string, maxDays int) error {
	logFilePath := filepath.Join(backupPath, "dbsnap.log")

	file, err := os.Open(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoffDate := time.Now().AddDate(0, 0, -maxDays)
	var validLines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Basic extraction assuming standard log.Ldate format (YYYY/MM/DD)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) > 0 {
			if logDate, err := time.Parse("2006/01/02", parts[0]); err == nil {
				if logDate.After(cutoffDate) {
					validLines = append(validLines, line)
				}
			} else {
				// Keep lines that don't match the exact date format just in case
				validLines = append(validLines, line)
			}
		}
	}
	file.Close()

	// Overwrite the file with only the valid lines
	return os.WriteFile(logFilePath, []byte(strings.Join(validLines, "\n")+"\n"), 0644)
}
