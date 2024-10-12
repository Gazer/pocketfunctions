// Run dart code and return a response
package languages

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/Gazer/pocketfunctions/models"
)

func DeployDartDocker(f *models.PocketFunction) (string, error) {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	cmd := exec.Command("docker", "build", "-t", f.Name, "--build-arg", fmt.Sprintf("name=%s", f.Name), ".")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Dir = "./docker_executor"

	if err := cmd.Run(); err != nil {
		log.Println(stdOut.String())
		log.Println(stdErr.String())
		return "", fmt.Errorf("Can not build docker image")
	}

	return StartDartDocker(f)
}

func StartDartDocker(f *models.PocketFunction) (string, error) {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	// Stop any other version of the container. Ignore errors
	log.Printf("Stopping container %s\n", f.DockerId)
	cmd := exec.Command("docker", "stop", f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Run()
	log.Println(stdOut.String())
	log.Println(stdErr.String())

	log.Printf("Removing container %s\n", f.DockerId)
	cmd = exec.Command("docker", "rm", f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Run()
	log.Println(stdOut.String())
	log.Println(stdErr.String())

	log.Printf("Startning new container %s\n", f.Name)
	port := fmt.Sprintf("%d:8080", f.Id+8080)
	cmd = exec.Command("docker", "run", "-p", port, "--name", f.Name, "-d", "--restart", "unless-stopped", f.Name)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Println(stdErr.String())
		return "", fmt.Errorf("Can not start docker image")
	}

	return stdOut.String(), nil
}
