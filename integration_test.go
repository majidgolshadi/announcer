package client_announcer

import (
	"flag"
	"os"
	"os/exec"
	"testing"
)

var integration = flag.Bool("integration", false, "run integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	if *integration {
		setupServices()
	}

	result := m.Run()
	if *integration {
		teardownServices()
	}
	os.Exit(result)
}

func setupServices() {
	exec.Command("docker-compose up -d")
}

func teardownServices() {
	exec.Command("docker-compose stop")
	exec.Command("docker-compose rm -f")
}
