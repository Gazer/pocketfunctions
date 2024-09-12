// Run dart code and return a response
package languages

import (
	"bytes"
	"errors"
	"fmt"
	"gin/models"
	"os"
	"os/exec"
	"strings"
)

var template = `%s`

func checkExecutorDirectory(dir string) {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Directory do not exist for %s ... creating one\n", dir)
		cmd := exec.Command("cp", "-a", "executor", dir)
		cmd.Run()
	}
}

func DeployDart(f *models.PocketFunction) {
	directory := fmt.Sprintf("./executors/%s", f.Id)
	testFile := fmt.Sprintf("./executors/%s/lib/file.dart", f.Id)

	checkExecutorDirectory(directory)

	executableCode := fmt.Sprintf(template, f.Code)
	if err := os.WriteFile(testFile, []byte(executableCode), 0666); err != nil {
		return
	}

	cmd := exec.Command("dart", "compile", "aot-snapshot", "bin/executor.dart")
	cmd.Dir = directory
	cmd.Run()
}

func RunDart(f *models.PocketFunction) (string, map[string]string, error) {
	directory := fmt.Sprintf("./executors/%s", f.Id)
	aotFile := fmt.Sprintf("./executors/%s/bin/executor.aot", f.Id)

	var headers map[string]string
	headers = make(map[string]string)
	var builder strings.Builder

	if _, err := os.Stat(aotFile); errors.Is(err, os.ErrNotExist) {
		// Code still not deployed
		fmt.Println("Deploying func ...")
		DeployDart(f)
	}

	cmd := exec.Command("dartaotruntime", "bin/executor.aot")
	cmd.Dir = directory

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", headers, errors.New("Command failed")
	}

	var lines = strings.Split(out.String(), "\n")
	var readingHeaders = true
	for _, line := range lines {
		if line == "====" {
			readingHeaders = false
		} else {
			if readingHeaders {
				var parts = strings.SplitN(line, "=", 2)
				headers[parts[0]] = parts[1]
			} else {
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
	}

	return builder.String(), headers, nil
}
