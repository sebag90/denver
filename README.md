<p align="center">
  <img src="assets/mark.svg" alt="Denver logo" width="140">
</p>

<h1 align="center">Denver</h1>

<p align="center">
  One development container per directory, managed by <code>dv</code>.
</p>

Denver is a small command-line tool that gives every project directory its own
long-lived container. Run a shell, execute commands, or leave it running in the
background — without writing a Compose file or rebuilding a devcontainer from
scratch.

It talks directly to the Podman or Docker-compatible API socket, so no CLI
wrapper or daemon of its own is involved.

## Features

- **Per-directory containers** — each working directory gets a stable container
  named `dv.<hash>`.
- **Persistent sessions** — the container runs an idle keep-alive process, so
  leaving a shell does not stop it and multiple shells can connect at once.
- **Podman-first** — uses Podman's libpod API when available, with rootless user
  mapping, and falls back to the Docker-compatible API.
- **Host conveniences** — mounts the current directory at the same absolute
  path, plus existing dotfiles, Wayland, and the container engine socket.
- **Optional sandbox mode** — skip the workspace bind and use the container as
  an isolated scratch environment.
- **Small and portable** — a single static Go binary for Linux, macOS, and
  FreeBSD.

## Requirements

- A Unix-like host (Linux, macOS, or FreeBSD).
- Podman or Docker with a reachable Unix API socket.
- Go 1.27.1 or newer, if building from source.

By default, `dv` looks for the rootless Podman socket at:

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

Prebuilt binaries for Linux, macOS, and FreeBSD (amd64 and arm64, plus
32-bit ARM on Linux) are attached to every
[release](https://github.com/sebag90/denver/releases). Each release publishes the
bare binaries `dv-<os>-<arch>`, e.g. `dv-linux-amd64` or `dv-darwin-arm64`.

### From a release

GitHub serves the assets of the newest non-prerelease release under a stable
`latest` path, so you can download just the binary that matches your host
without querying the API first:

```text
https://github.com/sebag90/denver/releases/latest/download/<asset-name>
```

Pick the asset for the current platform and install it:

```sh
case "$(uname -s)-$(uname -m)" in
  Linux-x86_64)              asset=dv-linux-amd64 ;;
  Linux-aarch64|Linux-arm64) asset=dv-linux-arm64 ;;
  Linux-armv6l|Linux-armv7l) asset=dv-linux-arm ;;
  Darwin-x86_64)             asset=dv-darwin-amd64 ;;
  Darwin-arm64)              asset=dv-darwin-arm64 ;;
  FreeBSD-x86_64)            asset=dv-freebsd-amd64 ;;
  FreeBSD-arm64)             asset=dv-freebsd-arm64 ;;
  *) echo "unsupported platform: $(uname -s)-$(uname -m)" >&2; exit 1 ;;
esac

curl -fsSL -o dv "https://github.com/sebag90/denver/releases/latest/download/$asset"
chmod +x dv
install -m 0755 dv ~/.local/bin/dv
```

To install a specific version, replace `latest` with the tag, for example
`releases/download/v0.1.0/dv-linux-amd64`. With the
[GitHub CLI](https://cli.github.com) you can also do:

```sh
gh release download --pattern 'dv-linux-amd64'   # add --tag v0.1.0 to pin
```

To list the available asset names instead of guessing them:

```sh
curl -fsSL https://api.github.com/repos/sebag90/denver/releases/latest \
  | jq -r '.assets[].browser_download_url'
```

Two caveats:

- Releases do not ship checksums, so `curl -f` over HTTPS is the only integrity
  check. Verify by building from source if that matters to you.
- The 32-bit `dv-linux-arm` build targets ARMv7 (Go's default `GOARM=7`); it
  will fault on older ARMv6 hosts.

### From source

Build the `dv` binary with [just](https://github.com/casey/just):

```sh
git clone https://github.com/sebag90/denver.git
cd denver
just build
install -m 0755 dv ~/.local/bin/dv
```

Or build it directly with Go:

```sh
go build -trimpath -ldflags="-s -w" -o dv .
```

Make sure the directory containing `dv` is on your `PATH`.

## Quick start

```sh
cd ~/src/my-project

dv up                  # create and start this directory's container
dv shell               # open an interactive login shell
dv run go test ./...   # run a one-off command
dv run -d ./fetch.sh   # start a job in the background
dv ps                  # list detached jobs
dv logs -f <job>       # follow a job's output
dv down                # stop and remove the container
dv down --all          # stop and remove the containers of every directory
dv list                # show all containers and their directories
```

The default image is `ghcr.io/sebag90/devenv:latest`. It is pulled
automatically the first time a container is created. Use another image with
the global `--image` flag:

```sh
dv --image ubuntu:24.04 shell
```

## Commands

| Command                      | Description                                                                                                                    |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `dv up`                      | Create and start the container for the current directory. Idempotent: an existing container is reused or restarted if stopped. |
| `dv shell`                   | Open an interactive TTY shell. Aliased as `dv ssh`. The host's `$SHELL` is tried first, with `/bin/sh` as fallback.            |
| `dv run <command> [args...]` | Run a command non-interactively and return its exit code. Flags after the command belong to the command.                       |
| `dv run -d <command> ...`    | Start a command in the background and return immediately. Output goes to a log file inside the container.                      |
| `dv ps`                      | List the detached jobs of the current directory's container. Aliased as `dv jobs`.                                             |
| `dv logs <job>`              | Print a detached job's output. With `-f`, follow it until the job ends.                                                        |
| `dv down`                    | Stop and remove the current directory's container. Other directories' containers are untouched.                                |
| `dv down --all`              | Stop and remove every container `dv` created, across all directories.                                                          |
| `dv list`                    | List every container with its directory, status, and whether it is sandboxed. Aliased as `dv ls`.                               |
| `dv help`                    | Show all commands and flags.                                                                                                   |
| `dv --version`               | Print the `dv` version.                                                                                                        |

### Flags

| Flag                   | Scope   | Description                                                                              |
| ---------------------- | ------- | ---------------------------------------------------------------------------------------- |
| `--image <ref>`        | global  | Image to use when creating a new container. Defaults to `ghcr.io/sebag90/devenv:latest`. |
| `-s`, `--sandbox`      | global  | Do not bind-mount the current directory into the container.                              |
| `-S`, `--shell <name>` | `shell` | Shell to start inside the container. Defaults to the basename of `$SHELL`, or `bash`.    |
| `--all`                | `down`  | Stop and remove the containers of every directory, not just the current one.              |
| `-d`, `--detach`       | `run`   | Start the command in the background instead of streaming its output.                      |
| `--log <path>`         | `run`   | Path inside the container for a detached command's output. Requires `--detach`.           |
| `-f`, `--follow`       | `logs`  | Keep printing new output until the job ends.                                              |

Flags are persistent, so `dv up --sandbox` and `dv --sandbox up` are
equivalent. For `dv run`, everything after the command name is passed through
unchanged:

```sh
dv run ls -la
dv run sh -c 'echo "hello from $HOSTNAME"'
```

### Detached jobs

`dv run --detach` starts a command and returns without waiting for it. This is
useful for long-running scripts, builds, or downloads:

```sh
dv run -d -- ./fetch-data.sh
dv ps                 # running, exited or lost, with exit code and age
dv logs -f <job>      # follow the output
```

A job's output and metadata live inside the container under
`/tmp/dv/jobs/<job>/`: `log` holds the output, `pid`, `started` and `cmd`
describe the process, and `exit` appears when it finishes. Use `--log <path>`
to write the output somewhere else inside the container, for example into the
bind-mounted workspace so it survives `dv down`:

```sh
dv run -d --log /work/project/fetch.log -- ./fetch-data.sh
```

Jobs run as normal exec sessions. They keep running while the container is up
and end when it is stopped, restarted, or removed; their metadata goes away
with the container. `dv ps` only knows about jobs started by `dv` (the engine
API has no way to list or re-attach to exec sessions), so use
`dv run ps aux` to see every process inside the container. There is no
interactive re-attach: use `dv logs -f` to watch output, or run the command
under `tmux` inside the container if you need to interact with it later.

## How it works

### Container identity

The container name is derived from the absolute path of the current directory
using an FNV-1a hash, and is prefixed with `dv.`. The same directory always
maps to the same container; different directories get different containers,
even if they share a name.

Older versions of `dv` named containers `denver.<hash>`. They are still
recognised: `dv down` removes one belonging to the current directory, and
`dv up` replaces it with a `dv.<hash>` container. `dv down --all` removes both
naming schemes.

The directory a container was created from is recorded in the
`denver.directory` label, and its sandbox setting in `denver.sandbox`, which is
what `dv list` shows. Containers created before the directory label existed are
still listed, but show `-` as their directory.

Because the name is path-based, moving or renaming a project directory makes
it look like a new project. Remove the old container with `dv down` before
moving, or remove the stale container manually.

### Mounts

By default, the project container includes:

| Host path          | Container path                                      | Notes                                                            |
| ------------------ | --------------------------------------------------- | ---------------------------------------------------------------- |
| Current directory  | Same absolute path                                  | Read-write, SELinux relabeled with `Z`. Disabled by `--sandbox`. |
| `~/.pi/agent`      | `/home/dev/.pi/agent`                               | Only when present.                                               |
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

Denver detects the engine by querying `/version` and `/info`:

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
just build       # build ./dv for the host
just build-all   # cross-compile into ./dist
just test        # go test ./...
just check       # tests plus go vet
just clean       # remove build artifacts
```

`just build-all` produces Linux, macOS, and FreeBSD binaries for amd64 and
arm64, plus Linux arm.

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
