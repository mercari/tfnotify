package localfile

import (
	"fmt"
	"os"
)

type OutputService service

const filePermission os.FileMode = 0o644

// WriteToFile Write result to file
func (f *OutputService) WriteToFile(body string, outputFile string) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermission)
	if err != nil {
		return fmt.Errorf("open a file to output the result to a file: %w", err)
	}

	defer file.Close()

	if _, err := file.WriteString(body + "\n"); err != nil {
		return fmt.Errorf("write the result to a file: %w", err)
	}
	return nil
}
