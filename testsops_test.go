package testsops

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v2"
)

type fakeConfig struct {
	FieldA int
	FieldB string
	FieldC fakeSubConfig
}

type fakeSubConfig struct {
	FieldD float64
	FieldE time.Time
}

// This is the fake loading function under test. We want to test that this
// function correctly decrypts valid sops-formatted files.
func loadFakeConfig(sopsEncFile string) (*fakeConfig, error) {
	format := strings.TrimPrefix(filepath.Ext(sopsEncFile), ".")
	clearText, err := decrypt.File(sopsEncFile, format)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}

	var cfg fakeConfig
	switch format {
	case "yml", "yaml":
		if err := yaml.Unmarshal(clearText, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	case "json":
		if err := json.Unmarshal(clearText, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case "env":
		return nil, errors.New("env file parsing is not implemented")
	}

	return &cfg, nil
}

func TestLoadFakeJSONConfig(t *testing.T) {
	rawJSON := `{
	"FieldA": 123,
	"FieldB": "abc",
	"FieldC: {
		"FieldD": 1.23,
		"FieldE": "2022-07-04T08:10:21.52Z"
	}
}`

	sopsCfg := EncryptJSON(t, rawJSON)
	t.SetEnv("SOPS_AGE_KEY_FILE", sopsCfg.KeyPath)
	got := loadFakeConfig(sopsCfg.EncryptedContentsPath)
	want := &fakeConfig{
		FieldA: 123,
		FieldB: "abc",
		FieldC: fakeSubConfig{
			FieldD: 1.23,
			FieldE: time.Unix(12345, 0),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected config returned: %v", err)
	}
}
