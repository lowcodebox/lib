#!/usr/bin/env bash
# scripts/run-drone-local.sh
set -euo pipefail
set -ex

drone exec --pipeline=lib --trusted --event=push --branch=feature/wpr-test