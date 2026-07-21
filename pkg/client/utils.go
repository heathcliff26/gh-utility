package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func newRequest(method string, url string, body any, token string) (req *http.Request, err error) {
	if body != nil {
		var buf []byte
		buf, err = json.Marshal(body)
		if err != nil {
			err = fmt.Errorf("failed to encode request body: %w", err)
			return
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(buf))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return
}

func fileToTreeObject(baseDir, path string) (*TreeObject, error) {
	obj := &TreeObject{
		Path: path,
		Type: treeTypeBlob,
		Mode: treeModeFile,
	}

	if baseDir != "" {
		path = filepath.Join(baseDir, path)
	}
	// #nosec G302 G304 -- File permissions and path are wanted this way
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if errors.Is(err, os.ErrNotExist) {
		obj.DeleteFile = true
		return obj, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to open file `%s`: %w", path, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get mode of file `%s`: %w", path, err)
	}
	if stat.Mode().Perm()&0001 == 1 {
		obj.Mode = treeModeExec
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file `%s`: %w", path, err)
	}
	obj.Content = string(buf)

	return obj, nil
}

// Convert the raw status to status and conclusion pair.
func translateCheckRunStatus(status string) (string, string, error) {
	lower := strings.ToLower(status)
	switch lower {
	case "succeeded", "success":
		return "completed", "success", nil
	case "failed", "failure":
		return "completed", "failure", nil
	case "skipped", "none":
		return "completed", "skipped", nil
	case "cancelled":
		return "completed", "cancelled", nil
	case "neutral":
		return "completed", "neutral", nil
	case "timed_out", "timeout":
		return "completed", "timed_out", nil
	case "action_required":
		return "completed", "action_required", nil
	case "stale":
		return "completed", "stale", nil
	case "running", "in_progress":
		return "in_progress", "", nil
	case "pending", "queued":
		return "queued", "", nil
	default:
		return "", "", fmt.Errorf("unknown status '%s'", status)
	}
}
