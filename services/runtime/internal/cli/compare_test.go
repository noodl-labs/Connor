package cli

import (
	"testing"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func TestLoadRunArtifact_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/run.json"
	artifact := entities.RunArtifact{
		Version: entities.RunArtifactVersion,
		SuiteID: "serving-smoke",
		Cases: []entities.RunCase{
			{ID: "a", Model: "m", Passed: true, LatencyMs: 100, Attempts: 1},
		},
		Summary: entities.RunSummary{Total: 1, Passed: 1, P95Ms: 100, P50Ms: 100, PassRate: 100},
	}
	if err := writeRunArtifact(path, artifact); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadRunArtifact(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SuiteID != "serving-smoke" || loaded.Summary.P95Ms != 100 {
		t.Fatalf("loaded: %+v", loaded)
	}
}
