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
    "FieldC": {
        "FieldD": 1.23,
        "FieldE": "2022-07-04T08:10:21.52Z"
    }
}`

	sopsCfg := EncryptJSON(t, rawJSON)
	t.Setenv("SOPS_AGE_KEY_FILE", sopsCfg.KeyPath)
	got, err := loadFakeConfig(sopsCfg.EncryptedContentsPath)
	if err != nil {
		t.Fatalf("loadFakeConfig: %v", err)
	}
	want := &fakeConfig{
		FieldA: 123,
		FieldB: "abc",
		FieldC: fakeSubConfig{
			FieldD: 1.23,
			FieldE: time.Unix(1656922221, 520000000).UTC(),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected config returned (-want +got)\n%s", diff)
	}
}

func TestLoadFakeYAMLConfig(t *testing.T) {
	rawYAML := `
fielda: 123
fieldb: "abc"
fieldc:
  fieldd: 1.23
  fielde: "2022-07-04T08:10:21.52Z"
`

	sopsCfg := EncryptYAML(t, rawYAML)
	t.Setenv("SOPS_AGE_KEY_FILE", sopsCfg.KeyPath)
	got, err := loadFakeConfig(sopsCfg.EncryptedContentsPath)
	if err != nil {
		t.Fatalf("loadFakeConfig: %v", err)
	}
	want := &fakeConfig{
		FieldA: 123,
		FieldB: "abc",
		FieldC: fakeSubConfig{
			FieldD: 1.23,
			FieldE: time.Unix(1656922221, 520000000).UTC(),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected config returned (-want +got)\n%s", diff)
	}
}
