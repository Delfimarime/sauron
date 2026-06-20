package docker

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	// gitImage ships both git and an OpenSSH server with a shell, so the seed
	// entrypoint can build a bare repo and serve it over ssh without a build step
	// or any runtime package install. It is the codebase-sanctioned git image.
	gitImage = "gitea/gitea:1"
	// gitUser is the account the binary clones as; gitRepo is the bare repository
	// path served over ssh. gitDefaultRef/gitPinnedRef are the branch the repo
	// checks out by default and the extra tag a scenario can pin via --ref.
	gitUser       = "git"
	gitRepoName   = "acme.git"
	gitDefaultRef = "main"
	gitPinnedRef  = "v1.0.0"

	// In-container locations the seed entrypoint and the binary's ssh client use.
	gitHome        = "/home/git"
	gitRepoPath    = gitHome + "/" + gitRepoName
	gitAuthKeys    = gitHome + "/.ssh/authorized_keys"
	gitHostKey     = "/etc/ssh/ssh_host_ed25519_key"
	gitEntrypoint  = "/seed.sh"
	gitSSHPort     = 22
	gitClientKey   = "/root/.ssh/id_ed25519"
	gitKnownHosts  = "/root/.ssh/known_hosts"
	gitSSHConfPath = "/root/.ssh/config"
)

// gitSource serves provider content from a bare git repository over ssh, served by
// an sshd sidecar. URL returns the ssh remote the binary clones; the seed
// entrypoint commits the exposed content on the default branch and also tags it as
// gitPinnedRef so a --ref scenario has a non-default ref to pin to.
type gitSource struct {
	resourceSet
	alias string
	keys  *sshKeyPair
	exec  serviceExec
}

// URL is the ssh remote the binary clones. Resolving it forces Start (the proxy
// guarantees that), at which point the sshd sidecar is up on the compose network.
func (s *gitSource) URL(context.Context) (string, error) {
	return fmt.Sprintf("ssh://%s@%s:%d%s", gitUser, gitService(s.alias), gitSSHPort, gitRepoPath), nil
}

func (s *gitSource) Path(context.Context) (string, error) {
	return "", fmt.Errorf("docker: git source %q has no path; use its url", s.alias)
}

// SSHKey is the in-container path of the mounted client private key the binary
// presents to authenticate against the sshd sidecar (gitClientMounts mounts it
// there). The clone verifies the sidecar's host key against the known_hosts entry
// mounted at the default location, so no host-key checking is disabled.
func (s *gitSource) SSHKey(context.Context) (string, error) {
	return gitClientKey, nil
}

// Revision is the commit the served repository's default branch resolves to,
// read by running git in the sshd sidecar. A scenario pins it via
// #{.git.<alias>.revision} to exercise commit-addressed resolution (a SHA is
// neither a branch nor a tag). safe.directory disarms the dubious-ownership guard,
// since the bare repo is owned by the git account but the exec runs as root.
func (s *gitSource) Revision(ctx context.Context) (string, error) {
	if s.exec == nil {
		return "", fmt.Errorf("docker: git source %q cannot resolve a revision", s.alias)
	}
	code, out, err := s.exec(ctx, gitService(s.alias),
		"git", "-c", "safe.directory=*", "-C", gitRepoPath, "rev-parse", gitDefaultRef)
	if err != nil {
		return "", fmt.Errorf("docker: resolve git revision for %q: %w", s.alias, err)
	}
	if code != 0 {
		return "", fmt.Errorf("docker: resolve git revision for %q: rev-parse exited %d: %s", s.alias, code, out)
	}
	return strings.TrimSpace(out), nil
}

// gitService is the compose service name (and in-network DNS host) of a git
// source's sshd sidecar.
func gitService(alias string) string { return "registry-git-" + alias }

// gitMounts turns a git source's content resources into mounts under the seed
// directory inside the sshd sidecar; the entrypoint commits them into the bare
// repo on startup.
func gitContentMounts(src *gitSource) []FileSpec {
	mounts := make([]FileSpec, 0, len(src.resources))
	for _, r := range src.resources {
		if r.IsAuth() {
			continue // git auth is key-based, not basic
		}
		mounts = append(mounts, FileSpec{Content: r.Content, Path: gitSeedDir + "/" + r.Path})
	}
	return mounts
}

const gitSeedDir = gitHome + "/seed"

// gitServerSpec builds the sshd sidecar that seeds and serves the bare repo. The
// generated public key is installed as authorized_keys and the matching host key
// is pinned, so the binary's clone authenticates without any interactive prompt.
func gitServerSpec(src *gitSource) ContainerSpec {
	mounts := make([]FileSpec, 0, len(src.resources)+3)
	mounts = append(mounts, gitContentMounts(src)...)
	mounts = append(mounts,
		FileSpec{Content: []byte(src.keys.AuthorizedKey + "\n"), Path: gitAuthKeys},
		FileSpec{Content: []byte(src.keys.HostPrivatePEM), Path: gitHostKey},
		FileSpec{Content: []byte(gitSeedScript()), Path: gitEntrypoint},
	)
	return ContainerSpec{
		Service:    gitService(src.alias),
		Image:      gitImage,
		Entrypoint: "sh " + gitEntrypoint,
		Mount:      mounts,
	}
}

// gitClientMounts are the files mounted into "main" so the binary can clone the
// remote: the private key, the pinned host key as a known_hosts entry, and an ssh
// config that selects the key and silences host-key prompts for the sidecar.
func gitClientMounts(src *gitSource) []FileSpec {
	host := gitService(src.alias)
	known := host + " " + strings.TrimSpace(src.keys.HostPublicLine)
	conf := strings.Join([]string{
		"Host " + host,
		"    User " + gitUser,
		"    IdentityFile " + gitClientKey,
		"    UserKnownHostsFile " + gitKnownHosts,
		"    StrictHostKeyChecking yes",
		"",
	}, "\n")
	return []FileSpec{
		{Content: []byte(src.keys.ClientPrivatePEM), Path: gitClientKey},
		{Content: []byte(known + "\n"), Path: gitKnownHosts},
		{Content: []byte(conf), Path: gitSSHConfPath},
	}
}

// gitSeedScript is the sidecar entrypoint: it builds a bare repo from the mounted
// seed content, commits it on the default branch, tags it as the pinned ref, wires
// authorized_keys, and execs sshd in the foreground. It uses the git binary the
// image already ships, so nothing is installed at runtime.
func gitSeedScript() string {
	return strings.Join([]string{
		"#!/bin/sh",
		"set -e",
		// Ensure the git account exists with our home and a real login shell,
		// independent of the base image's own layout.
		"id -u " + gitUser + " >/dev/null 2>&1 || adduser -D -h " + gitHome + " -s /bin/sh " + gitUser,
		// The base image's git account may declare a home (e.g. /data/git) that
		// does not exist; create whatever home it has so the ssh login can chdir.
		"mkdir -p \"$(awk -F: '$1==\"" + gitUser + "\"{print $6}' /etc/passwd)\"",
		"mkdir -p " + gitHome + "/.ssh /var/empty",
		"chmod 600 " + gitHostKey,
		"chmod 700 " + gitHome + "/.ssh",
		"chmod 600 " + gitAuthKeys,
		// Build the working tree from the seeded content, commit, tag, then push
		// into a bare repo the binary clones.
		"export HOME=" + gitHome,
		"git config --global --add safe.directory '*'",
		"git config --global user.email seed@example.com",
		"git config --global user.name seed",
		"git config --global init.defaultBranch " + gitDefaultRef,
		"git init --bare " + gitRepoPath,
		"work=$(mktemp -d)",
		"git init -b " + gitDefaultRef + " \"$work\"",
		"cp -r " + gitSeedDir + "/. \"$work\"/ 2>/dev/null || true",
		"cd \"$work\"",
		"git add -A",
		"git commit -m seed --allow-empty",
		"git tag " + gitPinnedRef,
		"git remote add origin " + gitRepoPath,
		"git push origin " + gitDefaultRef,
		"git push origin " + gitPinnedRef,
		"chown -R " + gitUser + " " + gitHome,
		// Serve over ssh in the foreground with our host key and authorized_keys,
		// not the image defaults (PID 1 keeps the container alive).
		"exec /usr/sbin/sshd -D -e" +
			" -o HostKey=" + gitHostKey +
			" -o AuthorizedKeysFile=" + gitAuthKeys +
			" -o PasswordAuthentication=no" +
			" -o PubkeyAuthentication=yes" +
			" -o UsePAM=no",
		"",
	}, "\n")
}

// sshKeyPair is a generated ed25519 client key plus the server host key, in the
// formats the fixture mounts: PEM private keys, an authorized_keys line for the
// client public key, and a known_hosts line for the host public key.
type sshKeyPair struct {
	ClientPrivatePEM string
	AuthorizedKey    string
	HostPrivatePEM   string
	HostPublicLine   string
}

// newSSHKeyPair generates a fresh client key (installed as authorized_keys) and a
// fresh host key (pinned in known_hosts), so each scenario authenticates without a
// prompt and without any host-side ssh tooling.
func newSSHKeyPair() (*sshKeyPair, error) {
	clientPriv, clientAuth, err := generateKey()
	if err != nil {
		return nil, fmt.Errorf("generate client key: %w", err)
	}
	hostPriv, hostPub, err := generateKey()
	if err != nil {
		return nil, fmt.Errorf("generate host key: %w", err)
	}
	return &sshKeyPair{
		ClientPrivatePEM: clientPriv,
		AuthorizedKey:    clientAuth,
		HostPrivatePEM:   hostPriv,
		HostPublicLine:   hostPub,
	}, nil
}

// generateKey returns one ed25519 key as an OpenSSH PEM private key and its public
// half in authorized_keys/known_hosts wire format.
func generateKey() (privatePEM, publicLine string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return "", "", err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", "", err
	}
	return string(pem.EncodeToMemory(block)), strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))), nil
}
