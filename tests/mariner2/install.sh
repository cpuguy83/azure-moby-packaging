#!/usr/bin/env bash

set -e

: ${TEST_ENGINE_PACKAGE_VERSION:=''}
: ${TEST_CLI_PACKAGE_VERSION:=''}
: ${TEST_CONTAINERD_PACKAGE_VERSION:=''}
: ${TEST_RUNC_PACKAGE_VERSION:=''}
: ${TEST_BUILDX_PACKAGE_VERSION:=''}
: ${TEST_COMPOSE_PACKAGE_VERSION:=''}

DEFAULT_REPO_DIR=/var/pkg

prepare_local_yum() {
    dir="${DEFAULT_REPO_DIR}"
    if [ -n "${1}" ]; then
        dir="${1}"
    fi
    createrepo "${dir}"
    dnf config-manager --add-repo "file://${dir}"
}

install() {
    tdnf install -y --nogpgcheck \
        moby-engine-"${TEST_ENGINE_PACKAGE_VERSION}*" \
        moby-cli-"${TEST_CLI_PACKAGE_VERSION}*" \
        moby-containerd-"${TEST_CONTAINERD_PACKAGE_VERSION}*" \
        moby-runc-"${TEST_RUNC_PACKAGE_VERSION}*" \
        moby-buildx-"${TEST_BUILDX_PACKAGE_VERSION}*" \
        moby-compose-"${TEST_COMPOSE_PACKAGE_VERSION}*"
}

init() {
    systemctl enable --now docker
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi
    systemctl enable --now docker.socket
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi

    systemctl enable --now containerd
    if [ ! $? -eq 0 ]; then
        journalctl -u containerd
        journalctl -xe
        exit 1
    fi

    systemctl start containerd
    if [ ! $? -eq 0 ]; then
        journalctl -u containerd
        journalctl -xe
        exit 1
    fi
    systemctl start docker
    if [ ! $? -eq 0 ]; then
        journalctl -u docker
        journalctl -xe
        exit 1
    fi
}

case "${1}" in
repo)
    prepare_local_yum "${2}"
    ;;
install)
    install
    init
    ;;
"")
    prepare_local_yum
    install
    init
    ;;
*)
    prepare_local_yum
    install
    init
    ;;
esac
