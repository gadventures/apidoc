#!/usr/bin/env bash

set -eu

go test -v -bench Benchmark -run Benchmark -count 3
