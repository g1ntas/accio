package gitgetter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var urlTests = []struct {
	name   string
	input  string
	raw    string
	subdir string
	ref    string
}{
	{"simple url", "http://repo.com/test", "http://repo.com/test", "", ""},
	{"url with subdirectory", "http://repo.com//subdir1/subdir2", "http://repo.com", "subdir1/subdir2", ""},
	{"url with ref", "http://repo.com#abc", "http://repo.com", "", "abc"},
	{"url with query param and ref", "http://repo.com?param=1#abc", "http://repo.com?param=1", "", "abc"},
	{"url with subdir and ref", "https://repo.com//subdir#abc", "https://repo.com", "subdir", "abc"},
	{"file url", "file:///path/to/repo", "file:///path/to/repo", "", ""},
	{"file url with subdir", "file:///path/to/repo//subdir", "file:///path/to/repo", "subdir", ""},

	// scp-style ssh
	{"scp ssh url", "user@host.com:repo", "ssh://user@host.com/repo", "", ""},
	{"scp ssh url with subdir", "user@host.com:repo//subdir1/subdir2", "ssh://user@host.com/repo", "subdir1/subdir2", ""},
	{"scp ssh url with ref", "user@host.com:repo#test", "ssh://user@host.com/repo", "", "test"},

	// github
	{"github scheme-less url", "github.com/owner/repo", "https://github.com/owner/repo", "", ""},
	{"github scheme-less basic auth url", "user:pass@github.com/owner/repo", "https://user:pass@github.com/owner/repo", "", ""},
	{"github url with subdir", "github.com/owner/repo/subdir1/subdir2", "https://github.com/owner/repo", "subdir1/subdir2", ""},
	{"github url with double-slash subdir", "github.com/owner/repo//subdir1/subdir2", "https://github.com/owner/repo", "subdir1/subdir2", ""},
	{"github ssh url with subdir", "ssh://git@github.com/owner/repo/subdir1/subdir2", "ssh://git@github.com/owner/repo", "subdir1/subdir2", ""},
	{"github scp url with subdir", "git@github.com:owner/repo/subdir1/subdir2", "ssh://git@github.com/owner/repo", "subdir1/subdir2", ""},
	{"github git url with subdir", "git://github.com/owner/repo/subdir1/subdir2", "git://github.com/owner/repo", "subdir1/subdir2", ""},

	// bitbucket
	{"bitbucket scheme-less url", "bitbucket.org/owner/repo", "https://bitbucket.org/owner/repo", "", ""},
	{"bitbucket scheme-less basic auth url", "user:pass@bitbucket.org/owner/repo", "https://user:pass@bitbucket.org/owner/repo", "", ""},
	{"bitbucket url with subdir", "bitbucket.org/owner/repo/subdir1/subdir2", "https://bitbucket.org/owner/repo", "subdir1/subdir2", ""},
	{"bitbucket url with double-slash subdir", "bitbucket.org/owner/repo//subdir1/subdir2", "https://bitbucket.org/owner/repo", "subdir1/subdir2", ""},
	{"bitbucket ssh url with subdir", "ssh://git@bitbucket.org/owner/repo/subdir1/subdir2", "ssh://git@bitbucket.org/owner/repo", "subdir1/subdir2", ""},
	{"bitbucket scp url with subdir", "git@bitbucket.org:owner/repo/subdir1/subdir2", "ssh://git@bitbucket.org/owner/repo", "subdir1/subdir2", ""},
	{"bitbucket git url with subdir", "git://bitbucket.org/owner/repo/subdir1/subdir2", "git://bitbucket.org/owner/repo", "subdir1/subdir2", ""},

	// gitlab
	{"gitlab scheme-less url", "gitlab.com/owner/repo", "https://gitlab.com/owner/repo", "", ""},
	{"gitlab scheme-less basic auth url", "user:pass@gitlab.com/owner/repo", "https://user:pass@gitlab.com/owner/repo", "", ""},
	{"gitlab url with subdir", "gitlab.com/owner/repo/subdir1/subdir2", "https://gitlab.com/owner/repo", "subdir1/subdir2", ""},
	{"gitlab url with double-slash subdir", "gitlab.com/owner/repo//subdir1/subdir2", "https://gitlab.com/owner/repo", "subdir1/subdir2", ""},
	{"gitlab ssh url with subdir", "ssh://git@gitlab.com/owner/repo/subdir1/subdir2", "ssh://git@gitlab.com/owner/repo", "subdir1/subdir2", ""},
	{"gitlab scp url with subdir", "git@gitlab.com:owner/repo/subdir1/subdir2", "ssh://git@gitlab.com/owner/repo", "subdir1/subdir2", ""},
	{"gitlab git url with subdir", "git://gitlab.com/owner/repo/subdir1/subdir2", "git://gitlab.com/owner/repo", "subdir1/subdir2", ""},

	// gitea
	{"gitea scheme-less url", "gitea.com/owner/repo", "https://gitea.com/owner/repo", "", ""},
	{"gitea scheme-less basic auth url", "user:pass@gitea.com/owner/repo", "https://user:pass@gitea.com/owner/repo", "", ""},
	{"gitea url with subdir", "gitea.com/owner/repo/subdir1/subdir2", "https://gitea.com/owner/repo", "subdir1/subdir2", ""},
	{"gitea url with double-slash subdir", "gitea.com/owner/repo//subdir1/subdir2", "https://gitea.com/owner/repo", "subdir1/subdir2", ""},
	{"gitea ssh url with subdir", "ssh://git@gitea.com/owner/repo/subdir1/subdir2", "ssh://git@gitea.com/owner/repo", "subdir1/subdir2", ""},
	{"gitea scp url with subdir", "git@gitea.com:owner/repo/subdir1/subdir2", "ssh://git@gitea.com/owner/repo", "subdir1/subdir2", ""},
	{"gitea git url with subdir", "git://gitea.com/owner/repo/subdir1/subdir2", "git://gitea.com/owner/repo", "subdir1/subdir2", ""},
}

func TestUrl(t *testing.T) {
	for _, test := range urlTests {
		t.Run(test.name, func(t *testing.T) {
			r, err := parseUrl(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.raw, r.raw)
			assert.Equal(t, test.subdir, r.subdir)
			assert.Equal(t, test.ref, r.ref)
		})
	}
}
