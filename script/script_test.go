package script

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/micahhausler/container-tx/compose"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestEmitContainers(t *testing.T) {
	cf := compose.DockerCompose{}

	f, err := os.Open("./test_fixtures/docker-compose.yaml")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	bp, err := cf.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

	got, err := Script{}.EmitContainers(bp)
	if err != nil {
		t.Errorf("Failed to emit containers: %s", err)
	}

	expected, err := ioutil.ReadFile("./test_fixtures/compose.out")
	if err != nil {
		t.Errorf("Failed to open file: %s", err)
	}

	if bytes.Compare(got, expected) != 0 {
		diff := diffmatchpatch.New()
		diffs := diff.DiffMain(string(expected), string(got), false)
		t.Errorf("Input differs from output: %s", diff.PatchToText(diff.PatchMake(diffs)))
	}
}
