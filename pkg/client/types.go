package client

import "encoding/json"

const (
	treeModeFile = "100644"
	treeModeExec = "100755"

	treeTypeBlob = "blob"
)

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type TreeObject struct {
	Path    string `json:"path"`
	Mode    string `json:"mode,omitempty"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (t *TreeObject) MarshalJSON() ([]byte, error) {
	var res any
	if t.Content == "" {
		tmp := &treeObjectWithSHA{
			Path: t.Path,
			Mode: t.Mode,
			Type: t.Type,
		}
		res = tmp
	} else {
		tmp := &treeObjectWithoutSHA{
			Path:    t.Path,
			Mode:    t.Mode,
			Type:    t.Type,
			Content: t.Content,
		}
		res = tmp
	}
	return json.Marshal(res)
}

type treeObjectWithSHA struct {
	Path string  `json:"path"`
	Mode string  `json:"mode"`
	Type string  `json:"type"`
	SHA  *string `json:"sha"`
}

type treeObjectWithoutSHA struct {
	Path    string `json:"path"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type TreeRequest struct {
	Tree     []*TreeObject `json:"tree"`
	BaseTree string        `json:"base_tree,omitempty"`
}

type TreeResponse struct {
	SHA string `json:"sha"`
}

type CommitRequest struct {
	Message string   `json:"message"`
	Tree    string   `json:"tree"`
	Parents []string `json:"parents,omitempty"`
}

type CommitResponse struct {
	SHA string `json:"sha"`
}

type BranchRequest struct {
	Ref   string `json:"ref,omitempty"`
	SHA   string `json:"sha"`
	Force *bool  `json:"force,omitempty"`
}
