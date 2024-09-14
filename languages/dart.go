// Run dart code and return a response
package languages

import (
	"archive/zip"
	"bufio"
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

func checkExecutorDirectory(dir string, f *models.PocketFunction) {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Directory do not exist for %s ... creating one\n", dir)
		cmd := exec.Command("cp", "-a", "executor", dir)
		cmd.Run()

		writeEntryPoint(f)
		writeDependencies(f)

		cmd1 := exec.Command("dart", "pub", "get")
		cmd1.Dir = dir
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
	os.Mkdir(fmt.Sprintf("./executors/%s/vendor", f.Id), 0755)

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
			outFile.WriteString(fmt.Sprintf("  %s:\n", f.Code))
			outFile.WriteString(fmt.Sprintf("    path: vendor/%s\n", f.Code))
			outFile.WriteString("\n")
		}
	}

	outFile.Close()
}

func Unzip(zipFilePath, destDir string) error {
	// Abrir el archivo ZIP
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo ZIP: %w", err)
	}
	defer zipReader.Close()

	// Iterar sobre los archivos en el archivo ZIP
	for _, file := range zipReader.File {
		// Construir la ruta completa para el archivo extraído
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			// Crear directorio si es un directorio
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("error al crear el directorio %s: %w", filePath, err)
			}
			continue
		}

		// Crear el archivo
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("error al crear el directorio del archivo %s: %w", filePath, err)
		}

		destFile, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error al crear el archivo %s: %w", filePath, err)
		}
		defer destFile.Close()

		// Copiar el contenido del archivo ZIP al archivo destino
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
	destDir := fmt.Sprintf("executors/%s/vendor/%s", f.Id, f.Code) // Reemplaza "yourVariable" con el valor real

	if err := Unzip(zipFilePath, destDir); err != nil {
		fmt.Println("Error descomprimiendo el archivo:", err)
	} else {
		fmt.Println("Archivo descomprimido exitosamente en:", destDir)
	}
}

func DeployDart(f *models.PocketFunction) {
	directory := fmt.Sprintf("./executors/%s", f.Id)

	checkExecutorDirectory(directory, f)
	unzipCode(f)

	vendorPath := fmt.Sprintf("executors/%s/vendor/%s", f.Id, f.Code)

	cmd := exec.Command("dart", "run", "build_runner", "build")
	cmd.Dir = vendorPath
	cmd.Run()

	cmd = exec.Command("dart", "pub", "get")
	cmd.Dir = directory
	cmd.Run()

	cmd = exec.Command("dart", "compile", "aot-snapshot", "bin/executor.dart")
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
