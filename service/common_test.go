package service

import (
	"bytes"
	"fmt"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"os"
	"os/exec"
	"strings"
)

var apiError404 = api.NewError(
	404, &sacloud.ResultErrorValue{
		ErrorCode:    "not_found",
		ErrorMessage: "The target can not be found. Object or state that can not be available, there is an error in the ID or path.",
		IsFatal:      true,
		Serial:       "xxx",
		Status:       "404 Not Found",
	})

type dummyAPI struct {
	dbAPI iaas.DatabaseAPI
}

func (c *dummyAPI) AuthStatus() (*sacloud.AuthStatus, error) {
	return nil, nil
}
func (c *dummyAPI) MariaDB() iaas.DatabaseAPI {
	return c.dbAPI
}
func (c *dummyAPI) PostgreSQL() iaas.DatabaseAPI {
	return c.dbAPI
}

func existsTestEnvVars(envVars ...string) (res bool) {
	for _, env := range envVars {
		v, ok := os.LookupEnv(env)
		if ok && v != "" {
			return true
		}
	}
	return
}

func startDocker(image string, envVars map[string]string,
	ports []int, cmd string) (cleanup func(), err error) {

	// clear envs
	dockerEnvs := []string{
		"DOCKER_TLS_VERIFY",
		"DOCKER_HOST",
		"DOCKER_CERT_PATH",
		"DOCKER_MACHINE_NAME",
	}
	for _, env := range dockerEnvs {
		os.Unsetenv(env) // nolint
	}

	args := []string{"run", "-d"}

	for k, v := range envVars {
		os.Setenv(k, v) // nolint
		args = append(args, "-e")
		args = append(args, k)
	}
	for _, p := range ports {
		args = append(args, "-p")
		args = append(args, fmt.Sprintf("127.0.0.1:%d:%d", p, p))
	}
	if cmd != "" {
		args = append(args, cmd)
	}
	args = append(args, image)

	buf := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	dockerCmd := exec.Command("docker", args...)
	dockerCmd.Stdout = buf
	dockerCmd.Stderr = stderr

	err = dockerCmd.Run()
	if err != nil {
		panic(fmt.Errorf("docker run is failed: %s\n%s", err, stderr))
	}

	cid := strings.TrimSpace(buf.String())
	cleanup = func() {
		cleanCmd := exec.Command("docker", "rm", "-f", cid)
		if e := cleanCmd.Run(); e != nil {
			panic(e)
		}
	}
	return
}
