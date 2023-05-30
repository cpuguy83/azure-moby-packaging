package git

import (
	"context"
	"os"
	"time"

	"dagger.io/dagger"
)

var knownHosts = `
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
`

// Container returns a container that can be used to run git commands.
// SSH_AUTH_SOCK is automatically forwarded to the container if it is set.
func Container(client *dagger.Client) *dagger.Container {
	ctr := client.Pipeline("git").Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "git", "openssh-client"})

	if authSock := os.Getenv("SSH_AUTH_SOCK"); authSock != "" {
		ctr = ctr.WithUnixSocket(authSock, client.Host().UnixSocket(authSock)).
			WithEnvVariable("SSH_AUTH_SOCK", authSock).
			WithNewFile("/root/.ssh/known_hosts", dagger.ContainerWithNewFileOpts{
				Contents:    knownHosts,
				Permissions: 0644,
			})
	}

	return ctr
}

// Fetch fetches the given ref from the given repo.
func Fetch(client *dagger.Client, repo, ref string) *dagger.Directory {
	var socket *dagger.Socket

	if authSock := os.Getenv("SSH_AUTH_SOCK"); authSock != "" {
		socket = client.Host().UnixSocket(authSock)
	}

	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref).Tree(dagger.GitRefTreeOpts{
		SSHKnownHosts: knownHosts,
		SSHAuthSocket: socket,
	})
}

// CommitTime gets the commit time of the given ref in the given repo.
func CommitTime(ctx context.Context, client *dagger.Client, repo, ref string) (time.Time, error) {
	out, err := Container(client).Pipeline("git commit time").
		WithMountedDirectory("/build/src", Fetch(client, repo, ref)).
		WithWorkdir("/build/src").
		WithExec([]string{
			"/bin/sh", "-ec",
			"date -u --date=$(git show -s --format=%cI HEAD) +%s",
		}).Stdout(ctx)

	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339, string(out))
	return t, err
}
