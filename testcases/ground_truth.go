package testcases

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Name             string    `yaml:"name"`
	Query            string    `yaml:"query"`
	RelevantChunks   []ChunkID `yaml:"relevant_chunks"`
	MinPrecisionAt5  float64   `yaml:"min_precision_at_5"`
	ExcludedPatterns []string  `yaml:"excluded_patterns,omitempty"`
}

type GroundTruth struct {
	TestCases           []TestCase `yaml:"testcases"`
	MinAveragePrecision float64    `yaml:"min_average_precision"`
	EvaluateTopK        int        `yaml:"evaluate_top_k"`
}

type ChunkID struct {
	FilePath   string   `yaml:"file_path"`
	HeaderPath []string `yaml:"header_path,omitempty"`
	Preview    string   `yaml:"preview"`
}

func LoadGroundTruth(path string) (*GroundTruth, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", path, err)
	}

	var gt GroundTruth

	err = yaml.Unmarshal(data, &gt)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	if gt.EvaluateTopK == 0 {
		gt.EvaluateTopK = 5
	}

	if gt.MinAveragePrecision == 0 {
		gt.MinAveragePrecision = 0.6
	}

	for i := range gt.TestCases {
		if gt.TestCases[i].MinPrecisionAt5 == 0 {
			gt.TestCases[i].MinPrecisionAt5 = gt.MinAveragePrecision
		}
	}
	return &gt, nil
}
