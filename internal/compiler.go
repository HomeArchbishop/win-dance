package internal

import (
	"os"
	"os/exec"
)

func Compile(outputBinary, targetFile string, cmdStrChan chan string) error {
	cmd := exec.Command("g++", targetFile, "-o", outputBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdStrChan <- cmd.String()
	close(cmdStrChan)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
