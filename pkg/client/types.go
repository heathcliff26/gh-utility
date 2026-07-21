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
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Content string `json:"content"`

	DeleteFile bool `json:"-"`
}

func (t *TreeObject) MarshalJSON() ([]byte, error) {
	var res any
	if t.DeleteFile {
		tmp := &struct {
			Path string  `json:"path"`
			Mode string  `json:"mode"`
			Type string  `json:"type"`
			SHA  *string `json:"sha"`
		}{
			Path: t.Path,
			Mode: t.Mode,
			Type: t.Type,
		}
		res = tmp
	} else {
		tmp := &struct {
			Path    string `json:"path"`
			Mode    string `json:"mode"`
			Type    string `json:"type"`
			Content string `json:"content"`
		}{
			Path:    t.Path,
			Mode:    t.Mode,
			Type:    t.Type,
			Content: t.Content,
		}
		res = tmp
	}
	return json.Marshal(res)
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

type PrRequest struct {
	Title string `json:"title,omitempty"`
	Head  string `json:"head,omitempty"`
	Base  string `json:"base,omitempty"`
	Body  string `json:"body,omitempty"`
}

type PrResponse struct {
	Url     string  `json:"url,omitempty"`
	Id      int     `json:"id,omitempty"`
	HtmlUrl string  `json:"html_url,omitempty"`
	Number  int     `json:"number,omitempty"`
	State   string  `json:"state,omitempty"`
	Title   string  `json:"title,omitempty"`
	Body    string  `json:"body,omitempty"`
	Labels  []Label `json:"labels,omitempty"`
	Head    struct {
		Label string `json:"label,omitempty"`
		Ref   string `json:"ref,omitempty"`
	} `json:"head,omitzero"`
	Base struct {
		Label string `json:"label,omitempty"`
		Ref   string `json:"ref,omitempty"`
	} `json:"base,omitzero"`
}

type LabelRequest struct {
	Labels []string `json:"labels"`
}

type Label struct {
	Name string `json:"name"`
}
