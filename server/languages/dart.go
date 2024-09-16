// Run dart code and return a response
package languages

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"gin/models"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var template = `%s`

func checkExecutorDirectory(f *models.PocketFunction) {
	if _, err := os.Stat(f.BasePath()); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Directory do not exist for %s ... creating one\n", f.BasePath())
		cmd := exec.Command("cp", "-a", "executor", f.BasePath())
		cmd.Run()

		writeEntryPoint(f)
		writeDependencies(f)

		cmd1 := exec.Command("dart", "pub", "get")
		cmd1.Dir = f.VendorPath()
		cmd1.Run()
	}
}

func writeEntryPoint(f *models.PocketFunction) {
	bytes, err := os.ReadFile(fmt.Sprintf("./executors/%s/bin/executor.dart.template", f.Id))
	if err != nil {
		fmt.Println("Can't read template")
		return
	}
	var template = string(bytes)

	var executorDart = fmt.Sprintf(template, f.Code)

	os.WriteFile(fmt.Sprintf("./executors/%s/bin/executor.dart", f.Id), []byte(executorDart), 0666)
}

func writeDependencies(f *models.PocketFunction) {
	f.MakeVendorPath()

	lines, _ := f.ReadPubspec()

	outFile, err := os.Create(f.PubspecPath())
	if err != nil {
		fmt.Println(err.Error() + `: ` + f.PubspecPath())
		return
	}

	for _, line := range lines {
		outFile.WriteString(line)
		outFile.WriteString("\n")
		if line == "dependencies:" {
			outFile.WriteString(fmt.Sprintf("  %s:\n", f.Code))
			outFile.WriteString(fmt.Sprintf("    path: vendor/%s\n", f.Code))
			outFile.WriteString("\n")
		}
	}

	outFile.Close()
}

func Unzip(zipFilePath, destDir string) error {
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo ZIP: %w", err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("error al crear el directorio %s: %w", filePath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("error al crear el directorio del archivo %s: %w", filePath, err)
		}

		destFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error al crear el archivo %s: %w", filePath, err)
		}
		defer destFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("error al abrir el archivo dentro del ZIP %s: %w", file.Name, err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("error al copiar el archivo %s: %w", filePath, err)
		}
	}

	return nil
}

func unzipCode(f *models.PocketFunction) {
	zipFilePath := fmt.Sprintf("function_repository/%s.zip", f.Code)

	if err := Unzip(zipFilePath, f.VendorPath()); err != nil {
		fmt.Println("Unzil failed")
	} else {
		fmt.Println("Unzil done")
	}
}

func DeployDart(f *models.PocketFunction) {
	checkExecutorDirectory(f)
	unzipCode(f)

	vendorPath := f.VendorPath()

	cmd := exec.Command("dart", "run", "build_runner", "build")
	cmd.Dir = vendorPath
	cmd.Run()

	cmd = exec.Command("dart", "pub", "get")
	cmd.Dir = f.BasePath()
	cmd.Run()

	cmd = exec.Command("dart", "compile", "aot-snapshot", "bin/executor.dart")
	cmd.Dir = f.BasePath()
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
	var stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr

	err := cmd.Run()

	if err != nil {
		return stdErr.String(), responseHeaders, errors.New("Command failed")
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
