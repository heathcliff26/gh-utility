package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTreeObjectMarshalJSON(t *testing.T) {
	t.Run("DeleteFile", func(t *testing.T) {
		require := require.New(t)

		tree := &TreeObject{
			Path:       "testpath",
			Mode:       "testmode",
			Type:       "testtype",
			DeleteFile: true,
		}
		buf, err := json.Marshal(tree)

		require.NoError(err)
		require.Equal(`{"path":"testpath","mode":"testmode","type":"testtype","sha":null}`, string(buf), "Should return expected json")
	})
	t.Run("Content", func(t *testing.T) {
		require := require.New(t)

		tree := &TreeObject{
			Path:    "testpath",
			Mode:    "testmode",
			Type:    "testtype",
			Content: "testcontent",
		}
		buf, err := json.Marshal(tree)

		require.NoError(err)
		require.Equal(`{"path":"testpath","mode":"testmode","type":"testtype","content":"testcontent"}`, string(buf), "Should return expected json")
	})
	t.Run("EmptyFile", func(t *testing.T) {
		require := require.New(t)

		tree := &TreeObject{
			Path:    "testpath",
			Mode:    "testmode",
			Type:    "testtype",
			Content: "",
		}
		buf, err := json.Marshal(tree)

		require.NoError(err)
		require.Equal(`{"path":"testpath","mode":"testmode","type":"testtype","content":""}`, string(buf), "Should return expected json")
	})
}
