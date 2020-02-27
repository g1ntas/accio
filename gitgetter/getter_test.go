package gitgetter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var gitSchemeRemoverTests = []struct {
	name        string
	url         string
	returnedUrl string
}{
	{"github with https", "https://github.com/test", "github.com/test"},
	{"github with http", "http://github.com/test", "github.com/test"},
	{"github with www and http", "http://www.github.com/test", "github.com/test"},
	{"github with www and https", "https://www.github.com/test", "github.com/test"},
	{"github without scheme", "github.com/test", "github.com/test"},

	{"bitbucket with https", "https://bitbucket.org/test", "bitbucket.org/test"},
	{"bitbucket with http", "http://bitbucket.org/test", "bitbucket.org/test"},
	{"bitbucket with www and http", "http://www.bitbucket.org/test", "bitbucket.org/test"},
	{"bitbucket with www and https", "https://www.bitbucket.org/test", "bitbucket.org/test"},
	{"bitbucket without scheme", "bitbucket.org/test", "bitbucket.org/test"},

	{"unknown url", "http://unknown.url", "http://unknown.url"},
	{"empty url", "", ""},
}

func TestGitSchemeRemover(t *testing.T) {
	for _, test := range gitSchemeRemoverTests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.returnedUrl, removeSchemeForGitServiceUrl(test.url))
		})
	}
}
