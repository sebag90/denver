# Build tasks for denver.
#
# The project is pure Go, so CGO is disabled and cross-compilation
# requires no extra toolchains.
#
# Windows is intentionally absent: internal/container/resize.go uses
# syscall.SIGWINCH, which does not exist on Windows.

binary  := "dv"
dist    := "dist"
targets := "linux/amd64 linux/arm64 linux/arm darwin/amd64 darwin/arm64 freebsd/amd64 freebsd/arm64"

# List the available recipes.
default:
    @just --list

# Build for the host platform into ./denver.
build:
    go build -trimpath -ldflags="-s -w" -o {{binary}} .

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
            go build -trimpath -ldflags="-s -w" \
            -o "{{dist}}/{{binary}}-${os}-${arch}" .
    done
    ls -lh {{dist}}

# Run the test suite.
test:
    go test ./...

# Run go vet and the tests.
check: test
    go vet ./...

# Install into GOBIN/GOPATH/bin.
install:
    go install -trimpath -ldflags="-s -w" .

# Remove build artifacts.
clean:
    rm -rf {{dist}} {{binary}}
