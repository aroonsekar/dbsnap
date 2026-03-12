package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var file *os.File

// Init sets up the dbsnap.log file in the target backup directory.
func Init(backupPath string) error {
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}

	logFilePath := filepath.Join(backupPath, "dbsnap.log")

	var err error
	file, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Configure standard logger to write to this file with dates and times
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime)

	return nil
}

// Close ensures the file buffer is flushed and closed after the job finishes.
func Close() {
	if file != nil {
		file.Close()
	}
}

// Info logs standard operational messages
func Info(format string, v ...any) {
	msg := fmt.Sprintf("[INFO] "+format, v...)
	log.Println(msg)
	fmt.Println(msg) // Also print to terminal if run manually
}

// Error logs failure states
func Error(format string, v ...any) {
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	log.Println(msg)
	fmt.Println(msg)
}
