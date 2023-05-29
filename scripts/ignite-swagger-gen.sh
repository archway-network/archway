#!/usr/bin/env bash

command -v ignite >/dev/null 2>&1 || { echo >&2 "❌ Require ignite-cli but it's not installed. Aborting swagger gen."; exit 1; }
ignite generate openapi -y