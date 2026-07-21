package status

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("SingleCheck", func(t *testing.T) {
		assert := assert.New(t)

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list check runs")
				assert.Equal("/repos/owner/repo/commits/commit-sha/check-runs", r.URL.Path, "Should use check-runs commit endpoint")

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"total_count":0,"check_runs":[]}`))
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should create check run")
				assert.Equal("/repos/owner/repo/check-runs", r.URL.Path, "Should use check-runs endpoint")

				var req client.CheckRun
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err, "Should decode check run request")
				assert.Equal("check-name", req.Name, "Should have correct name")
				assert.Equal("commit-sha", req.HeadSHA, "Should have correct head_sha")
				assert.Equal("completed", req.Status, "Should have translated status")
				assert.Equal("success", req.Conclusion, "Should have translated conclusion")
				assert.Equal("https://example.com", req.DetailsURL, "Should have details URL")
				assert.NotNil(req.Output, "Should have output")
				assert.Equal("check-name", req.Output.Title, "Should have correct output title")
				assert.Equal("Test Description", req.Output.Summary, "Should have correct output summary")

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1,"head_sha":"commit-sha","name":"check-name","node_id":"node","status":"completed","conclusion":"success"}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"check-name=succeeded",
		}
		flagArgs := []string{
			"--" + endpointFlag, s.URL,
			"--" + descriptionFlag, "Test Description",
			"--" + urlFlag, "https://example.com",
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd, positionalArgs)
		assert.NoError(err, "Should succeed")
		assert.Equal("Check run 'check-name' set to 'succeeded'\n", buf.String(), "Should print success message")
	})

	t.Run("UpdateExistingCheck", func(t *testing.T) {
		assert := assert.New(t)

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list check runs")
				assert.Equal("/repos/owner/repo/commits/commit-sha/check-runs", r.URL.Path)

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"total_count":1,"check_runs":[{"id":42,"head_sha":"commit-sha","name":"existing-check","node_id":"node","status":"queued","conclusion":""}]}`))
			case 1:
				assert.Equal(http.MethodPatch, r.Method, "Should update check run")
				assert.Equal("/repos/owner/repo/check-runs/42", r.URL.Path)

				var req client.CheckRun
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err, "Should decode check run request")
				assert.Equal("existing-check", req.Name, "Should have correct name")
				assert.Equal("completed", req.Status, "Should have translated status")
				assert.Equal("failure", req.Conclusion, "Should have translated conclusion")

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":42,"head_sha":"commit-sha","name":"existing-check","node_id":"node","status":"completed","conclusion":"failure"}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"existing-check=failed",
		}
		flagArgs := []string{
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd, positionalArgs)
		assert.NoError(err, "Should succeed")
		assert.Equal("Check run 'existing-check' set to 'failed'\n", buf.String(), "Should print success message")
	})

	t.Run("MultipleChecks", func(t *testing.T) {
		assert := assert.New(t)

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list check runs once")
				assert.Equal("/repos/owner/repo/commits/commit-sha/check-runs", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"total_count":0,"check_runs":[]}`))
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should create check-1")
				var req client.CheckRun
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err)
				assert.Equal("check-1", req.Name)
				assert.Equal("in_progress", req.Status)

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1,"head_sha":"commit-sha","name":"check-1","node_id":"node","status":"in_progress","conclusion":""}`))
			case 2:
				assert.Equal(http.MethodPost, r.Method, "Should create check-2")
				var req client.CheckRun
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err)
				assert.Equal("check-2", req.Name)
				assert.Equal("queued", req.Status)

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":2,"head_sha":"commit-sha","name":"check-2","node_id":"node","status":"queued","conclusion":""}`))
			case 3:
				assert.Equal(http.MethodPost, r.Method, "Should create check-3")
				var req client.CheckRun
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err)
				assert.Equal("check-3", req.Name)
				assert.Equal("completed", req.Status)
				assert.Equal("failure", req.Conclusion)

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":3,"head_sha":"commit-sha","name":"check-3","node_id":"node","status":"completed","conclusion":"failure"}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"check-1=running",
			"check-2=pending",
			"check-3=failed",
		}
		flagArgs := []string{
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd, positionalArgs)
		assert.NoError(err, "Should succeed")
		assert.Equal(4, requestCount, "Should make expected API calls")
		assert.Contains(buf.String(), "Check run 'check-1' set to 'running'", "Should report first check")
		assert.Contains(buf.String(), "Check run 'check-2' set to 'pending'", "Should report second check")
		assert.Contains(buf.String(), "Check run 'check-3' set to 'failed'", "Should report third check")
	})

	t.Run("InvalidCheckFormat", func(t *testing.T) {
		assert := assert.New(t)

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"invalidcheck",
		}
		flagArgs := []string{
			"--" + endpointFlag, "http://localhost",
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		err := run(cmd, positionalArgs)
		assert.Error(err, "Should fail with invalid check format")
		assert.Contains(err.Error(), "invalid check format", "Should mention invalid format")
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		assert := assert.New(t)

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list check runs")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"total_count":0,"check_runs":[]}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"check-name=unknown_status",
		}
		flagArgs := []string{
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		err := run(cmd, positionalArgs)
		assert.Error(err, "Should fail with unknown status")
		assert.Contains(err.Error(), "unknown status", "Should mention unknown status")
	})

	t.Run("ApiError", func(t *testing.T) {
		assert := assert.New(t)

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list check runs")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message":"Bad Request"}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		positionalArgs := []string{
			"owner/repo",
			"commit-sha",
			"check-name=success",
		}
		flagArgs := []string{
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(flagArgs)

		err := run(cmd, positionalArgs)
		assert.Error(err, "Should fail when API returns error")
		assert.Contains(err.Error(), "failed to get check runs for commit", "Should wrap API error")
	})
}
