#!/usr/bin/env bash

set -eo pipefail

# Check input
target_dir="$1"
if [ -z "${target_dir}" ]; then
  echo "Usage: $0 [target_dir]"
  exit 1
fi

# Add DONTCOVER
echo "Prepending '// DONTCOVER' to proto-generated files:"
find "${target_dir}" -type f -name '*.pb.go' -o -iname '*.pb.gw.go' | while read fname; do
  echo "${fname}"
  echo -e "// DONTCOVER\n$(cat ${fname})" > "${fname}"
done
