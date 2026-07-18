package entities

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNotComparable             = errors.New("entities: runs are not comparable")
	ErrUnknownRunArtifactVersion = errors.New("entities: unknown run.json version")
)

// P95Driver identifies the case that drives candidate suite p95 (UX on compare FAIL).
type P95Driver struct {
	Found        bool
	CaseID       string
	Model        string
	BaselineMs   int64
	CandidateMs  int64
	DeltaPercent float64
}

// P95CompareResult is the outcome of the p95 regression gate (RFC 0001 §5.2).
type P95CompareResult struct {
	Checked      bool
	Passed       bool
	DeltaPercent float64
	Threshold    float64
	BaselineP95  int64
	CandidateP95 int64
	Driver       P95Driver
}

// PassRateCompareResult is the outcome of the pass-rate gate (RFC 0001 §5.3).
type PassRateCompareResult struct {
	Checked           bool
	Passed            bool
	Threshold         float64
	CandidatePassRate float64
}

// CompareResult aggregates compare gates. Passed is true when all enabled gates pass.
type CompareResult struct {
	Passed   bool
	P95      P95CompareResult
	PassRate PassRateCompareResult
}

// ParseRunArtifactJSON decodes run.json and validates schema version.
func ParseRunArtifactJSON(data []byte) (RunArtifact, error) {
	var artifact RunArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return RunArtifact{}, fmt.Errorf("entities: decode run.json: %w", err)
	}
	if artifact.Version != RunArtifactVersion {
		return RunArtifact{}, fmt.Errorf("%w: got %d", ErrUnknownRunArtifactVersion, artifact.Version)
	}
	return artifact, nil
}

// ValidateComparable enforces ADR 0001: same suite_id, case ids, and models per case.
func ValidateComparable(baseline, candidate RunArtifact) error {
	if baseline.SuiteID != candidate.SuiteID {
		return fmt.Errorf("%w: suite_id %q vs %q", ErrNotComparable, baseline.SuiteID, candidate.SuiteID)
	}
	if len(baseline.Cases) != len(candidate.Cases) {
		return fmt.Errorf(
			"%w: case count %d vs %d",
			ErrNotComparable, len(baseline.Cases), len(candidate.Cases),
		)
	}
	for i := range baseline.Cases {
		bc := baseline.Cases[i]
		cc := candidate.Cases[i]
		if bc.ID != cc.ID {
			return fmt.Errorf(
				"%w: case id %q vs %q at index %d",
				ErrNotComparable, bc.ID, cc.ID, i,
			)
		}
		if bc.Model != cc.Model {
			return fmt.Errorf(
				"%w: model %q vs %q for case %q",
				ErrNotComparable, bc.Model, cc.Model, bc.ID,
			)
		}
	}
	return nil
}

// P95RegressionPercent returns percent change candidate vs baseline (RFC 0001 §5.2).
func P95RegressionPercent(baselineP95, candidateP95 int64) float64 {
	if baselineP95 == 0 {
		if candidateP95 == 0 {
			return 0
		}
		return 100
	}
	return float64(candidateP95-baselineP95) / float64(baselineP95) * 100
}

// FindP95Driver returns the case whose candidate latency matches suite p95 (or the slowest case).
func FindP95Driver(baseline, candidate RunArtifact) P95Driver {
	if len(candidate.Cases) == 0 {
		return P95Driver{}
	}

	target := candidate.Summary.P95Ms
	idx := -1
	for i, cc := range candidate.Cases {
		if cc.LatencyMs == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		for i, cc := range candidate.Cases {
			if idx < 0 || cc.LatencyMs > candidate.Cases[idx].LatencyMs {
				idx = i
			}
		}
	}

	baselineMs := baseline.Cases[idx].LatencyMs
	candidateMs := candidate.Cases[idx].LatencyMs
	return P95Driver{
		Found:        true,
		CaseID:       candidate.Cases[idx].ID,
		Model:        candidate.Cases[idx].Model,
		BaselineMs:   baselineMs,
		CandidateMs:  candidateMs,
		DeltaPercent: P95RegressionPercent(baselineMs, candidateMs),
	}
}

// CompareRuns compares two artifacts.
// maxP95Regression nil skips the p95 gate; minPassRate nil skips the pass-rate gate.
func CompareRuns(
	baseline, candidate RunArtifact,
	maxP95Regression *float64,
	minPassRate *float64,
) (CompareResult, error) {
	if err := ValidateComparable(baseline, candidate); err != nil {
		return CompareResult{}, err
	}

	delta := P95RegressionPercent(baseline.Summary.P95Ms, candidate.Summary.P95Ms)
	result := CompareResult{
		Passed: true,
		P95: P95CompareResult{
			DeltaPercent: delta,
			BaselineP95:  baseline.Summary.P95Ms,
			CandidateP95: candidate.Summary.P95Ms,
		},
		PassRate: PassRateCompareResult{
			CandidatePassRate: candidate.Summary.PassRate,
		},
	}

	if maxP95Regression != nil {
		result.P95.Checked = true
		result.P95.Threshold = *maxP95Regression
		result.P95.Passed = delta <= *maxP95Regression
		if !result.P95.Passed {
			result.Passed = false
			result.P95.Driver = FindP95Driver(baseline, candidate)
		}
	}

	if minPassRate != nil {
		result.PassRate.Checked = true
		result.PassRate.Threshold = *minPassRate
		result.PassRate.Passed = candidate.Summary.PassRate >= *minPassRate
		if !result.PassRate.Passed {
			result.Passed = false
		}
	}

	return result, nil
}
