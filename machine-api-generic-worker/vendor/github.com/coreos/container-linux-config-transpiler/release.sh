#!/usr/bin/env bash

set -e

[ $# == 1 ] || { echo "usage: $0 version" && exit 1; }

VER=$1

[[ "${VER}" =~ ^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+$ ]] || {
	echo "malformed version: \"${VER}\""
	exit 2
}

if test -n "$(git ls-files --others | \
	grep --invert-match '\(gopath/src/github.com/coreos/container-linux-config-transpiler\|bin/ct\)')";
then
	echo "directory has untracked files"
	exit 1
fi

if ! $(git diff-index --quiet HEAD --); then
	echo "directory has uncommitted files"
	exit 1
fi

make test

git tag --sign --message "Config-Transpiler ${VER}" "${VER}"

git verify-tag --verbose "${VER}"

make release
