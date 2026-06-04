package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
