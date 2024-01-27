#!/usr/bin/env bash

set -xe

cd "$(mktemp -d)"
curl -LO "https://github.com/fboulnois/pg_uuidv7/releases/download/v1.4.1/{pg_uuidv7.tar.gz,SHA256SUMS}"
tar xf pg_uuidv7.tar.gz
sha256sum -c SHA256SUMS
PG_MAJOR=$(pg_config --version | sed 's/^.* \([0-9]\{1,\}\).*$/\1/')
cp "$PG_MAJOR/pg_uuidv7.so" "$(pg_config --pkglibdir)"
cp pg_uuidv7--1.4.sql pg_uuidv7.control "$(pg_config --sharedir)/extension"

sudo -u postgres psql -c "CREATE EXTENSION IF NOT EXISTS pg_uuidv7;"
