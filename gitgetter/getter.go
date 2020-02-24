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
			"git": new(getter.GitGetter),
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
	switch {
	case strings.HasPrefix(src, "http://"):
		src = src[7:]
	case strings.HasPrefix(src, "https://"):
		src = src[8:]
	default:
		return src
	}
	if strings.HasPrefix(src, "www.") {
		src = src[4:]
	}
	if !strings.HasPrefix(src, "github.com/") && !strings.HasPrefix(src, "bitbucket.org/") {
		return src
	}
	return src
}
