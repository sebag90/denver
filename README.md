<p align="center">
  <img src="assets/mark.svg" alt="Paguro logo" width="140">
</p>

<h1 align="center">Paguro</h1>

<p align="center">
  One development container per directory, managed by <code>paguro</code>
  (or <code>pg</code>).
</p>

Paguro is a small command-line tool that gives every project directory its own
long-lived container. Run a shell, execute commands, or leave it running in the
background — without writing a Compose file or rebuilding a devcontainer from
scratch.

## Features

- **Per-directory containers** — each working directory gets a stable container
  named `paguro.<hash>`.
- **Persistent sessions** — the container runs an idle keep-alive process, so
  leaving a shell does not stop it and multiple shells can connect at once.
- **Podman-first** — uses Podman's libpod API when available, with rootless user
  mapping, and falls back to the Docker-compatible API.
- **Small and portable** — a single static Go binary for Linux, macOS, and
  FreeBSD.

## Requirements

- A Unix-like host (Linux, macOS, or FreeBSD).
- Podman or Docker with a reachable Unix API socket.
- Go 1.27.1 or newer, if building from source.

By default, `paguro` looks for the rootless Podman socket at:

```text
$XDG_RUNTIME_DIR/podman/podman.sock
```

or `/run/user/<uid>/podman/podman.sock` when `XDG_RUNTIME_DIR` is unset. Set
`CONTAINER_HOST` or `DOCKER_HOST` to a `unix://` URL to use another socket:

```sh
export CONTAINER_HOST=unix:///run/user/1000/podman/podman.sock
# or
export DOCKER_HOST=unix:///var/run/docker.sock
```

On macOS, the Podman socket lives inside the Podman machine, so point
`CONTAINER_HOST` at the socket reported by `podman machine inspect`.

Windows is not supported.

## Installation

The install script picks the right asset for the current platform and puts it in
`~/.local/bin` as `paguro`, with a `pg` symlink beside it:

```sh
curl -fsSL https://raw.githubusercontent.com/sebag90/paguro/main/install.sh | sh
# pin a version
curl -fsSL https://raw.githubusercontent.com/sebag90/paguro/main/install.sh | sh -s -- v0.3.1
# install somewhere else
curl -fsSL https://raw.githubusercontent.com/sebag90/paguro/main/install.sh | PAGURO_INSTALL_DIR=~/bin sh
```

### From source

Build the `paguro` binary with [just](https://github.com/casey/just):

```sh
git clone https://github.com/sebag90/paguro.git
cd paguro
just build
install -m 0755 paguro ~/.local/bin/paguro
ln -sf paguro ~/.local/bin/pg
```

Or build it directly with Go:

```sh
go build -trimpath -ldflags="-s -w" -o paguro .
```

Make sure the directory containing `paguro` is on your `PATH`.

## Quick start

```sh
cd ~/src/my-project

paguro up                  # create and start this directory's container
paguro shell               # open an interactive login shell
paguro run go test ./...   # run a one-off command
paguro run -d ./fetch.sh   # start a job in the background
paguro ps                  # list detached jobs
paguro ps --all            # list detached jobs of every running container
paguro logs -f <job>       # follow a job's output
paguro down                # stop and remove the container
paguro down --all          # stop and remove the containers of every directory
paguro list                # show all containers and their directories
```

`paguro` and `pg` are the same binary, so every command above also works with
`pg` (for example `pg up` or `pg shell`).

The default image is [`ghcr.io/sebag90/devenv:latest`](https://github.com/sebag90/devenv/pkgs/container/devenv). It is pulled
automatically the first time a container is created. Use another image with
the global `--image` flag:

```sh
paguro --image ubuntu:24.04 shell
```

## Commands

| Command                            | Description                                                                                                                          |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `paguro up`                        | Create and start the container for the current directory. Idempotent: an existing container is reused or restarted if stopped.       |
| `paguro shell`                     | Open an interactive TTY shell. Aliased as `paguro ssh`. The host's `$SHELL` is tried first, with `/bin/sh` as fallback.              |
| `paguro run <command> [args...]`   | Run a command non-interactively and return its exit code. Flags after the command belong to the command.                             |
| `paguro run -d <command> ...`      | Start a command in the background and return immediately. Output goes to a log file inside the container.                            |
| `paguro ps`                        | List the detached jobs of the current directory's container, with the directory each job runs in. Aliased as `paguro jobs`.          |
| `paguro ps --all`                  | List the detached jobs of every running container. Jobs of stopped containers cannot be read.                                        |
| `paguro logs <job>`                | Print a detached job's output, wherever its container lives. With `-f`, follow it until the job ends.                                |
| `paguro down`                      | Stop and remove the current directory's container. Other directories' containers are untouched.                                      |
| `paguro down --all`                | Stop and remove every container `paguro` created, across all directories.                                                            |
| `paguro list`                      | List every container with its directory, status, and whether it is sandboxed. Aliased as `paguro ls`.                                |
| `paguro help`                      | Show all commands and flags.                                                                                                         |
| `paguro --version`                 | Print the `paguro` version.                                                                                                          |

### Flags

| Flag                   | Scope   | Description                                                                              |
| ---------------------- | ------- | ---------------------------------------------------------------------------------------- |
| `--image <ref>`        | global  | Image to use when creating a new container. Defaults to `ghcr.io/sebag90/devenv:latest`. |
| `-s`, `--sandbox`      | global  | Do not bind-mount the current directory into the container.                              |
| `-S`, `--shell <name>` | `shell` | Shell to start inside the container. Defaults to the basename of `$SHELL`, or `bash`.    |
| `--all`                | `down`  | Stop and remove the containers of every directory, not just the current one.             |
| `--all`                | `ps`    | List the jobs of every running container, not just the current directory's.              |
| `-d`, `--detach`       | `run`   | Start the command in the background instead of streaming its output.                     |
| `--log <path>`         | `run`   | Path inside the container for a detached command's output. Requires `--detach`.          |
| `-f`, `--follow`       | `logs`  | Keep printing new output until the job ends.                                             |

### Detached jobs

`paguro run --detach` starts a command and returns without waiting for it. This
is useful for long-running scripts, builds, or downloads:

```sh
paguro run -d -- ./fetch-data.sh
paguro ps                 # running, exited or lost, with exit code and age
paguro ps --all           # same, for every running container
paguro logs -f <job>      # follow the output
```

A job's output and metadata live inside the container under
`/tmp/paguro/jobs/<job>/`: `log` holds the output, `pid`, `started` and `cmd`
describe the process, and `exit` appears when it finishes. Use `--log <path>`
to write the output somewhere else inside the container, for example into the
bind-mounted workspace so it survives `paguro down`:

```sh
paguro run -d --log /work/project/fetch.log -- ./fetch-data.sh
```

Jobs run as normal exec sessions. They keep running while the container is up
and end when it is stopped, restarted, or removed; their metadata goes away
with the container. `paguro ps` only knows about jobs started by `paguro` (the
engine API has no way to list or re-attach to exec sessions), so use
`paguro run ps aux` to see every process inside the container. There is no
interactive re-attach: use `paguro logs -f` to watch output, or run the command
under `tmux` inside the container if you need to interact with it later.

`paguro ps` lists the jobs of the current directory's container. `paguro ps
--all` walks every running container `paguro` created and shows a `DIRECTORY`
column, so background work can be found from anywhere. Jobs of a stopped
container are not listed: `paguro` reads their metadata by running a command
inside the container, and a stopped container cannot be entered.

`paguro logs <job>` searches the current directory's container first and then
every other running container, so a job listed by `paguro ps --all` can be read
without `cd`ing to its directory. That relies on job IDs being unique: they are
12 random hex characters (48 bits), which makes a collision between two
containers improbable rather than impossible.

## How it works

### Container identity

The container name is derived from the absolute path of the current directory
using an FNV-1a hash, and is prefixed with `paguro.`. The same directory always
maps to the same container; different directories get different containers,
even if they share a name.

The directory a container was created from is recorded in the
`paguro.directory` label, and its sandbox setting in `paguro.sandbox`, which is
what `paguro list` shows. Containers that do not carry the directory label
still show up, but print `-` as their directory.

Because the name is path-based, moving or renaming a project directory makes
it look like a new project. Remove the old container with `paguro down` before
moving, or remove the stale container manually.

### Mounts

By default, the project container includes:

| Host path          | Container path                                      | Notes                                                            |
| ------------------ | --------------------------------------------------- | ---------------------------------------------------------------- |
| Current directory  | Same absolute path                                  | Read-write, SELinux relabeled with `Z`. Disabled by `--sandbox`. |
| `~/.pi/agent`      | `/home/dev/.pi/agent`                               | Only when present.                                               |
| `~/.config/pig`    | `/home/dev/.config/pig`                             | Only when present.                                               |
| `~/.aws`           | `/home/dev/.aws`                                    | Only when present.                                               |
| `~/.gitconfig`     | `/home/dev/.gitconfig`                              | Only when present.                                               |
| `~/.ssh`           | `/home/dev/.ssh`                                    | Only when present.                                               |
| `~/.netrc`         | `/home/dev/.netrc`                                  | Only when present.                                               |
| `$WAYLAND_DISPLAY` | `/tmp/wayland-0`                                    | Only when a Wayland socket exists.                               |
| Engine socket      | `/run/podman/podman.sock` or `/var/run/docker.sock` | Mounted when available.                                          |

The engine socket is exposed so that `podman` or `docker` can be used _inside_
the project container. Containers started there run on the host daemon, next to
the project container itself (docker-outside-of-docker).

The container runs as the host UID/GID. Rootless Podman additionally uses
`keep-id`, so files created in bind mounts stay owned by the host user.

> **Security note:** the container can read the mounted SSH keys, cloud
> credentials, `.netrc`, and the engine socket. Treat it as trusted code
> execution. `--sandbox` removes only the workspace mount; it does not remove
> the credential or socket mounts.

### Engine differences

Paguro detects the engine by querying `/version` and `/info`:

- **Podman** is created through the libpod API, which allows a passwd entry for
  the host user and proper rootless user-namespace mapping.
- **Docker-compatible engines** use the standard `/containers/create` API.

When a container needs access to a host socket (the engine socket or Wayland),
SELinux labeling is disabled for that container. Relabeling a host socket
would break the host session. Containers without socket mounts keep their
SELinux confinement.

API versions are negotiated with the daemon automatically, so older engines
are not asked for an API version they do not speak.

## Building and testing

```sh
just build       # build ./paguro (and the ./pg alias) for the host
just build-all   # cross-compile into ./dist
just test        # go test ./...
just check       # tests plus go vet
just clean       # remove build artifacts
```

`just build-all` produces Linux, macOS, and FreeBSD binaries for amd64 and
arm64, plus Linux arm.

## Releasing

Releases are built by the [release workflow](.github/workflows/release.yml),
triggered manually from GitHub:

1. Make sure the commit you want to release is on `main`.
2. Go to **Actions → Release → Run workflow**, choose `main`, and enter the
   version (for example `v0.4.0`).
3. The workflow runs the tests, cross-compiles every target with that version
   compiled into the binary, creates the `v0.4.0` tag at the workflow's commit,
   and publishes a GitHub release with the binaries and `SHA256SUMS`.

Enable **dry_run** to run the tests and builds without creating a tag or
release; the binaries remain available as workflow artifacts. If a run fails
after creating the draft release, delete the draft (and its tag) before
retrying, since the workflow refuses to overwrite an existing tag or release.

Pre-release versions such as `v0.4.0-rc.1` are published as GitHub
pre-releases.

## Project layout

```text
.
├── cmd/                  # Cobra commands: up, shell, run, ps, logs, down, list
├── internal/container/   # API client, container lifecycle, mounts, exec, pull
├── assets/               # Logo assets
├── justfile              # Build and test recipes
└── main.go               # Entry point
```

## License

See [LICENSE](LICENSE).
