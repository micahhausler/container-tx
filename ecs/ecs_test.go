package ecs

import (
	"os"
	"testing"
)

func TestIngestContainers(t *testing.T) {
	pod := Task{}

	f, err := os.Open("./test_fixtures/task.json")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	_, err = pod.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}
}

func TestEmitContainers(t *testing.T) {
	pod := Task{}

	f, err := os.Open("./test_fixtures/task.json")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	bp, err := pod.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

	_, err = Task{}.EmitContainers(bp)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

}
