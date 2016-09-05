package ecs

import (
	"testing"
	"os"
)

func TestIngestContainers(t *testing.T){
	pod := EcsFormat{}

	f, err := os.Open("./test_fixtures/task.json")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	_, err = pod.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}
}

func TestEmitContainers(t *testing.T){
	pod := EcsFormat{}

	f, err := os.Open("./test_fixtures/task.json")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	bp, err := pod.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

	_, err = EcsFormat{}.EmitContainers(bp)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

}