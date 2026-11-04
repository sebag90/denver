# Build tasks for paguro.
#
# The project is pure Go, so CGO is disabled and cross-compilation
# requires no extra toolchains.
#
# Windows is intentionally absent: internal/container/resize.go uses
# syscall.SIGWINCH, which does not exist on Windows.

binary  := "paguro"
alias   := "pg"
dist    := "dist"
targets := "linux/amd64 linux/arm64 linux/arm darwin/amd64 darwin/arm64 freebsd/amd64 freebsd/arm64"
pkg     := "github.com/sebag90/paguro"
version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
ldflags := "-s -w -X " + pkg + "/cmd.version=" + version

# List the available recipes.
default:
    @just --list

# Build for the host platform into ./paguro (with a ./pg alias).
build:
    podman run --rm \
        -v "{{justfile_directory()}}:/src:Z" -w /src \
        -v paguro-go-cache:/go -e GOCACHE=/go/cache \
        golang:1.27 \
        go build -trimpath -ldflags="{{ldflags}}" -o {{binary}} .
    ln -sf {{binary}} {{alias}}

# Build for every supported OS/arch into ./dist.
build-all:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p {{dist}}
    for target in {{targets}}; do
        os="${target%/*}"
        arch="${target#*/}"
        echo "==> ${os}/${arch}"
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
            go build -trimpath -ldflags="{{ldflags}}" \
            -o "{{dist}}/{{binary}}-${os}-${arch}" .
    done
    ls -lh {{dist}}

# Run the test suite.
test:
    go test ./...

# Run go vet and the tests.
check: test
    go vet ./...

# Install into GOBIN/GOPATH/bin, with a pg alias next to it.
install:
    #!/usr/bin/env bash
    set -euo pipefail
    go install -trimpath -ldflags="{{ldflags}}" .
    bindir="$(go env GOBIN)"
    [ -n "$bindir" ] || bindir="$(go env GOPATH)/bin"
    ln -sf {{binary}} "$bindir/{{alias}}"

# Remove build artifacts.
clean:
    rm -rf {{dist}} {{binary}} {{alias}}
