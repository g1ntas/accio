package gitgetter

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"net/url"
	"strings"
)

type repo struct {
	raw    string
	ref    string
	subdir string
}

var knownHosts = [...]string{
	"bitbucket.org",
	"github.com",
	"gitlab.com",
	"gitea.com",
}

func parseUrl(rawurl string) (repo, error) {
	r := repo{}
	rawurl = strings.TrimSpace(rawurl)
	if rawurl == "" {
		return r, errors.New("empty url")
	}
	if !hasSupportedUrlScheme(rawurl) {
		rawurl = detectAndNormalizeScpSshUrl(rawurl)
	}
	rawurl = detectAndNormalizeKnownHostUrl(rawurl)
	u, err := url.Parse(rawurl)
	if err != nil {
		return r, err
	}
	if l := strings.SplitN(u.Path, "//", 2); len(l) == 2 {
		u.Path = l[0]
		r.subdir = l[1]
	}
	r.ref = u.Fragment
	u.Fragment = ""
	r.raw = u.String()
	return r, nil
}

func detectAndNormalizeKnownHostUrl(rawurl string) string {
	if rawurl == "" {
		return rawurl
	}
	var newurl string
	var scheme string
	switch {
	case strings.HasPrefix(rawurl, "ssh://"):
		newurl = rawurl[6:]
		scheme = "ssh"
	case strings.HasPrefix(rawurl, "git://"):
		newurl = rawurl[6:]
		scheme = "git"
	default:
		newurl = strings.TrimPrefix(rawurl, "http://")
		newurl = strings.TrimPrefix(newurl, "https://")
		newurl = strings.TrimPrefix(newurl, "www.")
		scheme = "https"
	}
	var auth string
	if i := strings.Index(newurl, "@"); i > -1 {
		auth = newurl[:i+1]
		newurl = newurl[i+1:]
	}
	for _, host := range knownHosts {
		if !strings.HasPrefix(newurl, host+"/") {
			continue
		}
		if strings.Contains(newurl, "//") {
			return scheme + "://" + auth + newurl
		}
		parts := strings.SplitN(newurl, "/", 4)
		if len(parts) < 4 {
			return scheme + "://" + auth + newurl
		}
		return fmt.Sprintf("%s://%s%s/%s/%s//%s", scheme, auth, parts[0], parts[1], parts[2], parts[3])

	}
	return rawurl
}

func detectAndNormalizeScpSshUrl(rawurl string) string {
	scannedUser := !strings.ContainsRune(rawurl, '@')
	for i, r := range rawurl {
		switch {
		case 'A' <= r && r <= 'Z':
		case 'a' <= r && r <= 'z':
		case '0' <= r && r <= '9':
		case r == '_':
		case scannedUser && (r == '-' || r == '.'):
		case r == '@':
			scannedUser = true
		case scannedUser && r == ':':
			return fmt.Sprintf("ssh://%s/%s", rawurl[:i], rawurl[i+1:])
		default:
			return rawurl
		}
	}
	return rawurl
}

func hasSupportedUrlScheme(rawurl string) bool {
	rawurl = strings.ToLower(rawurl)
	for scheme := range client.Protocols {
		if strings.HasPrefix(rawurl, scheme+"://") {
			return true
		}
	}
	return false
}
