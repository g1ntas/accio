package gitgetter

import (
	"context"
	"github.com/hashicorp/go-getter"
	"strings"
)

type Client struct {
	Pwd string
}

func (c *Client) CloneRepository(ctx context.Context, src, dst string) error {
	gitGetter := new(getter.GitGetter)
	client := &getter.Client{
		Ctx:     ctx,
		Src:     removeSchemeForGitServiceUrl(src),
		Dst:     dst,
		Pwd:     c.Pwd,
		Mode:    getter.ClientModeDir,
		Options: []getter.ClientOption{},
		Detectors: []getter.Detector{
			new(getter.GitHubDetector),
			new(getter.BitBucketDetector),
			new(getter.GitDetector),
			new(ForcedGitDetector),
		},
		Getters: map[string]getter.Getter{
			"git": gitGetter,
			"http": gitGetter,
			"https": gitGetter,
		},
	}
	return client.Get()
}

// ForcedGitDetector forces to use git getter.
type ForcedGitDetector struct{}

func (d *ForcedGitDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}
	return "git::" + src, true, nil
}

// removeSchemeForGitServiceUrl removes url scheme/protocol part, if url points to github or bitbucket.
func removeSchemeForGitServiceUrl(src string) string {
	if len(src) == 0 {
		return src
	}
	var uri string
	switch {
	case strings.HasPrefix(src, "http://"):
		uri = src[7:]
	case strings.HasPrefix(src, "https://"):
		uri = src[8:]
	default:
		return src
	}
	if strings.HasPrefix(uri, "www.") {
		uri = uri[4:]
	}
	if !strings.HasPrefix(uri, "github.com/") && !strings.HasPrefix(uri, "bitbucket.org/") {
		return src
	}
	return uri
}
