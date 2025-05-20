package server

import (
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

func TestServerIntegration(t *testing.T) {
	s := NewAPIServerStub()

	// Start the server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", s); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(time.Minute * 30)

	// Use kubectl to interact with the server.
	// 1. Run kubectl get pods
	out, err := runKubectl("http://localhost:8080", "get", "pods")
	if err != nil {
		t.Fatalf("Failed to get pods: %v, out:\n %s", err, out)
	}
	t.Logf("kubectl get pods output:\n%s", out)
}

func runKubectl(server string, params ...string) (string, error) {
	args := []string{
		"--server", server,
	}
	args = append(args, params...)
	out, err := exec.Command("kubectl", args...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("failed to run kubectl: %w", err)
	}

	return string(out), nil
}
