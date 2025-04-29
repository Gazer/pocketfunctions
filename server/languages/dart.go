// Run dart code and return a response
package languages

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"

	"github.com/Gazer/pocketfunctions/models"
)

type DartLanguage struct {
    f *models.PocketFunction
}

func newDart(f *models.PocketFunction) *DartLanguage {
    dart := DartLanguage{f: f}

    return &dart
}

func (dart *DartLanguage) CopyFile(file multipart.File) (error) {
	dst, err := os.Create(fmt.Sprintf("docker_executor/dist/%s.zip", dart.f.Name))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}
    return nil
}

func (dart *DartLanguage) Deploy() (string, error) {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	cmd := exec.Command("docker", "build", "-t", dart.f.Name, "--build-arg", fmt.Sprintf("name=%s", dart.f.Name), ".")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Dir = "./docker_executor"

	if err := cmd.Run(); err != nil {
		log.Println(stdOut.String())
		log.Println(stdErr.String())
		return "", fmt.Errorf("Can not build docker image")
	}

	return dart.Start()
}

func (dart *DartLanguage) Start() (string, error) {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	// Stop any other version of the container. Ignore errors
	log.Printf("Stopping container %s\n", dart.f.DockerId)
	cmd := exec.Command("docker", "stop", dart.f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Run()
	log.Println(stdOut.String())
	log.Println(stdErr.String())

	log.Printf("Removing container %s\n", dart.f.DockerId)
	cmd = exec.Command("docker", "rm", dart.f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Run()
	log.Println(stdOut.String())
	log.Println(stdErr.String())

	log.Printf("Startning new container %s\n", dart.f.Name)
	port := fmt.Sprintf("%d:8080", dart.f.Id+8080)
	cmd = exec.Command("docker", "run", "-p", port, "--name", dart.f.Name, "-d", "--restart", "unless-stopped", dart.f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Println(stdErr.String())
		return "", fmt.Errorf("Can not start docker image")
	}

	return stdOut.String(), nil
}
