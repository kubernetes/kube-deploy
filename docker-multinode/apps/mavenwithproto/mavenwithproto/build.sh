#!/usr/bin/env sh

set -e
apt-get update && apt-get install -y --no-install-recommends \
		build-base \
		autoconf \
		automake \
                libtool \
	&& rm -rf /var/lib/apt/lists/*

git clone https://github.com/google/protobuf -b $PROTOBUF_TAG --depth 1

cd ./protobuf

./autogen.sh && \
  ./configure --prefix=/usr && \
  make -j 3 && \
  make check && \
  make install

cd ..
rm -rf ./protobuf
