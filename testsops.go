package testsops

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
)

type Config struct {
	KeyPath               string
	EncryptedContentsPath string
}

type options struct {
	// sopsPath is the file path to the `sops` executable to use. Defaults to
	// looking in $PATH with exec.LookPath.
	sopsPath string
}

type Option func(*options)

// WithSOPSBinary specifies the path of the `sops` executable to use for
// encryption. If unspecified, defaults to looking in $PATH with exec.LookPath.
func WithSOPSBinary(sopsPath string) Option {
	return func(o *options) {
		o.sopsPath = sopsPath
	}
}

func EncryptFile(t *testing.T, fn string, opts ...Option) Config {
	ext := filepath.Ext(fn)
	if ext == "" {
		t.Fatalf("file %q had no extension, sops won't be able to detect format", fn)
	}
	if !strings.HasPrefix(fn, ".") {
		t.Fatalf("unexpected extension %q found on file", fn)
	}
	ext = strings.TrimPrefix(ext, ".")

	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatalf("failed to read file %q for encryption: %v", err)
	}
	return generateEncryptedConfig(t, string(dat), ext, opts...)
}

func EncryptYAML(t *testing.T, contents string, opts ...Option) Config {
	return generateEncryptedConfig(t, contents, "yaml", opts...)
}

func EncryptJSON(t *testing.T, contents string, opts ...Option) Config {
	return generateEncryptedConfig(t, contents, "json", opts...)
}

func EncryptEnv(t *testing.T, contents string, opts ...Option) Config {
	return generateEncryptedConfig(t, contents, "env", opts...)
}

func EncryptIni(t *testing.T, contents string, opts ...Option) Config {
	return generateEncryptedConfig(t, contents, "ini", opts...)
}

// generateEncryptedConfig takes the given contents and puts them into an
// encrypted sops file. In a real environment, we use KMS keys stored in GCP,
// but for testing, we use age [1], which is supported in sops and has great Go
// library support. This returns a configuration that contains the temp path of
// the encrypted config, and the path to the private key required to decrypt
// it.
// [1] https://pkg.go.dev/filippo.io/age#GenerateX25519Identity
func generateEncryptedConfig(t *testing.T, contents, ext string, opts ...Option) Config {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	tmpDir := t.TempDir()

	// Generate a new age identity.
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate 'age' identify: %v", err)
	}

	// Save the private key to a tmp file.
	keyPath := filepath.Join(tmpDir, "key.txt")
	if err := ioutil.WriteFile(keyPath, []byte(identity.String()), 0o700); err != nil {
		t.Fatalf("failed to save 'age' private key to tmp file: %v", err)
	}

	sopsPath := o.sopsPath
	if sopsPath == "" {
		if sopsPath, err = exec.LookPath("sops"); err != nil {
			t.Fatalf("failed to find 'sops' binary in path: %v", err)
		}
	}

	f, err := ioutil.TempFile(tmpDir, "contents-*.enc."+ext)
	if err != nil {
		t.Fatalf("failed to open temp file for encrypted contents: %v", err)
	}
	defer f.Close()

	// There's no sops library to do encryption, only decryption, so we have to
	// shell out to Sops for our test encryption.
	cmd := exec.Command(sopsPath, "--encrypt", "--age", identity.Recipient().String())
	cmd.Stdin = strings.NewReader(contents)
	cmd.Stderr = os.Stderr
	cmd.Stdout = f
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to encrypt contents with sops: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("failed to write encrypted contents file: %v", err)
	}

	return Config{
		KeyPath:               keyPath,
		EncryptedContentsPath: f.Name(),
	}
}
