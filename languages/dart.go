// Run dart code and return a response
package languages

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"gin/models"
	"os"
	"os/exec"
	"strings"
)

var template = `%s`

func checkExecutorDirectory(dir string, f *models.PocketFunction) {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Directory do not exist for %s ... creating one\n", dir)
		cmd := exec.Command("cp", "-a", "executor", dir)
		cmd.Run()

		writeDependencies(f)

		cmd1 := exec.Command("dart", "pub", "get")
		cmd1.Dir = dir
		cmd1.Run()
	}
}

func writeDependencies(f *models.PocketFunction) {
	yamlFile := fmt.Sprintf("./executors/%s/pubspec.yaml", f.Id)

	inFile, err := os.Open(yamlFile)
	if err != nil {
		fmt.Println(err.Error() + `: ` + yamlFile)
		return
	}

	var lines []string

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	inFile.Close()

	outFile, err := os.Create(yamlFile)
	if err != nil {
		fmt.Println(err.Error() + `: ` + yamlFile)
		return
	}

	for _, line := range lines {
		outFile.WriteString(line)
		outFile.WriteString("\n")
		if line == "dependencies:" {
			var lines = strings.Split(f.Deps, "\n")
			for _, dep := range lines {
				outFile.WriteString("  ")
				outFile.WriteString(dep)
				outFile.WriteString("\n")
			}
			outFile.WriteString("\n")
		}
	}

	outFile.Close()
}

func DeployDart(f *models.PocketFunction) {
	directory := fmt.Sprintf("./executors/%s", f.Id)
	testFile := fmt.Sprintf("./executors/%s/lib/file.dart", f.Id)

	checkExecutorDirectory(directory, f)

	executableCode := fmt.Sprintf(template, f.Code)
	if err := os.WriteFile(testFile, []byte(executableCode), 0666); err != nil {
		return
	}

	cmd := exec.Command("dart", "compile", "aot-snapshot", "bin/executor.dart")
	cmd.Dir = directory
	cmd.Run()
}

func RunDart(f *models.PocketFunction, env map[string]string) (string, map[string]string, error) {
	directory := fmt.Sprintf("./executors/%s", f.Id)
	aotFile := fmt.Sprintf("./executors/%s/bin/executor.aot", f.Id)

	var responseHeaders map[string]string
	responseHeaders = make(map[string]string)
	var builder strings.Builder

	if _, err := os.Stat(aotFile); errors.Is(err, os.ErrNotExist) {
		// Code still not deployed
		fmt.Println("Deploying func ...")
		DeployDart(f)
	}

	cmd := exec.Command("dartaotruntime", "bin/executor.aot")
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Dir = directory

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return out.String(), responseHeaders, errors.New("Command failed")
	}

	var lines = strings.Split(out.String(), "\n")
	var readingHeaders = true
	for _, line := range lines {
		if line == "====" {
			readingHeaders = false
		} else {
			if readingHeaders {
				var parts = strings.SplitN(line, "=", 2)
				responseHeaders[parts[0]] = parts[1]
			} else {
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
	}

	return builder.String(), responseHeaders, nil
}
