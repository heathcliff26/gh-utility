package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

const (
	UnchangedFile = "unchanged.txt"
	ChangedFile   = "changed.txt"
	StagedFile    = "staged.txt"
	DeletedFile   = "deleted.txt"
)

// Generate an RSA key.
// Returns the path to the keyfile.
func GenerateRSAKey(t *testing.T) string {
	require := require.New(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err, "Should generate RSA key")

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	path := filepath.Join(t.TempDir(), "private-key.pem")
	err = os.WriteFile(path, pemdata, 0600)
	require.NoError(err, "Should write RSA key")

	return path
}

func NewTestRepository(t *testing.T) (string, *git.Repository) {
	require := require.New(t)

	dir := t.TempDir()

	r, err := git.PlainInit(dir, false)
	require.NoError(err, "Should create new repository")

	for _, name := range []string{UnchangedFile, ChangedFile, StagedFile, DeletedFile} {
		// #nosec G306 -- Normal file permissions
		err = os.WriteFile(filepath.Join(dir, name), []byte("Placeholder text"), 0644)
		require.NoErrorf(err, "Should create new file %s", name)
	}

	tree, err := r.Worktree()
	require.NoError(err, "Should get worktree")
	require.NoError(tree.AddGlob("."), "Should stage files")
	author := object.Signature{
		Name:  "nobody",
		Email: "noreply@example.com",
		When:  time.Now(),
	}
	_, err = tree.Commit("Initial Commit", &git.CommitOptions{
		Author: &author,
	})
	require.NoError(err, "Should commit files")

	return dir, r
}
