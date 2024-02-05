#!/usr/bin/env bash
export GO_VERSION=${1:-"1.20.11"}
export ARCH=${2:-"amd64"}

# Download version
wget https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz

# Clear old installations and extract downloaded tarball
rm -rf /usr/local/go && tar -C /usr/local -xvzf go${GO_VERSION}.linux-${ARCH}.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
