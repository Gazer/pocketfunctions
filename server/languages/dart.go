// Run dart code and return a response
package languages

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Gazer/pocketfunctions/models"
)

func checkExecutorDirectory(f *models.PocketFunction) {
	if _, err := os.Stat(f.BasePath()); errors.Is(err, os.ErrNotExist) {
		log.Printf("Directory do not exist for %s ... creating one\n", f.BasePath())
		cmd := exec.Command("cp", "-a", "executor", f.BasePath())
		cmd.Run()

		writeEntryPoint(f)
		writeDependencies(f)

		cmd1 := exec.Command("dart", "pub", "get")
		cmd1.Dir = f.CodePath()
		cmd1.Run()
	}
}

func writeEntryPoint(f *models.PocketFunction) {
	bytes, err := os.ReadFile(fmt.Sprintf("../dist/executors/%d/bin/executor.dart.template", f.Id))
	if err != nil {
		log.Println("Can't read template")
		return
	}
	var executorDart = fmt.Sprintf(string(bytes), f.Code)

	os.WriteFile(fmt.Sprintf("../dist/executors/%d/bin/executor.dart", f.Id), []byte(executorDart), 0666)
}

func writeDependencies(f *models.PocketFunction) {
	os.Mkdir(f.VendorPath(), 0755)

	// Update executor pubspec to include the Code lib as dependency.
	lines, _ := f.ReadPubspec()

	outFile, err := os.Create(f.PubspecPath())
	if err != nil {
		log.Println(err.Error() + `: ` + f.PubspecPath())
		return
	}
	defer outFile.Close()

	for _, line := range lines {
		outFile.WriteString(line)
		outFile.WriteString("\n")
		if line == "dependencies:" {
			outFile.WriteString(fmt.Sprintf("  %s:\n", f.Code))
			outFile.WriteString(fmt.Sprintf("    path: vendor/%s\n", f.Code))
			outFile.WriteString("\n")
		}
	}
}

func unzip(zipFilePath, destDir string) error {
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("Can't open the ZIP: %w", err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("Can not create directory %s: %w", filePath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}

		destFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("Cant create file %s: %w", filePath, err)
		}
		defer destFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("Cant read the ZIP %s: %w", file.Name, err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("Cant copy file %s: %w", filePath, err)
		}
	}

	return nil
}

func unzipCode(f *models.PocketFunction) {
	zipFilePath := fmt.Sprintf("../dist/function_repository/%s.zip", f.Code)

	if err := unzip(zipFilePath, f.CodePath()); err != nil {
		log.Println("Unzil failed")
	} else {
		log.Println("Unzil done")
	}
}

func DeployDart(f *models.PocketFunction) {
	checkExecutorDirectory(f)
	unzipCode(f)

	cmd := exec.Command("dart", "run", "build_runner", "build")
	cmd.Dir = f.CodePath()
	cmd.Run()

	cmd = exec.Command("dart", "pub", "get")
	cmd.Dir = f.BasePath()
	cmd.Run()

	cmd = exec.Command("dart", "compile", "aot-snapshot", "bin/executor.dart")
	cmd.Dir = f.BasePath()
	cmd.Run()
}

func RunDart(f *models.PocketFunction, filePath string) (string, map[string]string, error) {
	aotFile := fmt.Sprintf("../dist/executors/%d/bin/executor.aot", f.Id)

	var responseHeaders map[string]string = make(map[string]string)
	var builder strings.Builder

	if _, err := os.Stat(aotFile); errors.Is(err, os.ErrNotExist) {
		// Code still not deployed
		log.Println("Deploying func ...")
		DeployDart(f)
	}

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	cmd := exec.Command("dartaotruntime", "bin/executor.aot", filePath)
	cmd.Dir = f.BasePath()
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()

	if err != nil {
		return stdErr.String(), responseHeaders, errors.New("Command failed")
	}

	var lines = strings.Split(stdOut.String(), "\n")
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
