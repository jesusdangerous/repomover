package git

import (
	"fmt"
	neturl "net/url"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type AuthConfig struct {
	Platform         string
	Token            string
	SSHUser          string
	SSHKeyPath       string
	SSHKeyPassphrase string
}

func authForRemote(remoteURL string, cfg AuthConfig, requireAuth bool) (transport.AuthMethod, string, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	platform := strings.ToLower(strings.TrimSpace(cfg.Platform))
	token := strings.TrimSpace(cfg.Token)

	if isSSHRemote(remoteURL) {
		auth, err := sshAuth(cfg)
		if err != nil {
			return nil, "ssh", err
		}
		return auth, "ssh", nil
	}

	if platform == "" {
		platform = detectPlatformFromURL(remoteURL)
	}

	if token == "" {
		switch platform {
		case "github":
			token = strings.TrimSpace(valueOrEnv("", "GITHUB_TOKEN"))
		case "gitlab":
			token = strings.TrimSpace(valueOrEnv("", "GITLAB_TOKEN"))
		}
	}

	if token == "" {
		if requireAuth {
			return nil, "none", fmt.Errorf("token is not set: use --token or GITHUB_TOKEN/GITLAB_TOKEN")
		}
		return nil, "none", nil
	}

	httpAuth, err := httpAuthForPlatform(platform, token)
	if err != nil {
		return nil, "http", err
	}

	return httpAuth, "http-token", nil
}

func httpAuthForPlatform(platform string, token string) (transport.AuthMethod, error) {
	switch platform {
	case "github", "":
		return &githttp.BasicAuth{Username: "token", Password: token}, nil
	case "gitlab":
		return &githttp.BasicAuth{Username: "oauth2", Password: token}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func sshAuth(cfg AuthConfig) (transport.AuthMethod, error) {
	sshUser := strings.TrimSpace(valueOrEnv(cfg.SSHUser, "REPOMOVER_SSH_USER"))
	if sshUser == "" {
		sshUser = "git"
	}

	keyPath := strings.TrimSpace(valueOrEnv(cfg.SSHKeyPath, "REPOMOVER_SSH_KEY_PATH"))
	keyPassphrase := strings.TrimSpace(valueOrEnv(cfg.SSHKeyPassphrase, "REPOMOVER_SSH_KEY_PASSPHRASE"))

	if keyPath != "" {
		auth, err := gitssh.NewPublicKeysFromFile(sshUser, keyPath, keyPassphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to load SSH key from REPOMOVER_SSH_KEY_PATH: %w", err)
		}
		return auth, nil
	}

	auth, err := gitssh.NewSSHAgentAuth(sshUser)
	if err != nil {
		return nil, fmt.Errorf("ssh auth is not configured: set SSH agent or REPOMOVER_SSH_KEY_PATH")
	}

	return auth, nil
}

func valueOrEnv(value, envKey string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return os.Getenv(envKey)
}

func isSSHRemote(remoteURL string) bool {
	remoteURL = strings.TrimSpace(remoteURL)
	lower := strings.ToLower(remoteURL)

	if strings.HasPrefix(lower, "ssh://") {
		return true
	}

	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") && !strings.Contains(remoteURL, "://") {
		return true
	}

	if parsed, err := neturl.Parse(remoteURL); err == nil {
		return strings.EqualFold(parsed.Scheme, "ssh")
	}

	return false
}

func detectPlatformFromURL(remoteURL string) string {
	raw := strings.ToLower(strings.TrimSpace(remoteURL))

	if strings.Contains(raw, "gitlab") {
		return "gitlab"
	}

	if strings.Contains(raw, "github") {
		return "github"
	}

	return ""
}
