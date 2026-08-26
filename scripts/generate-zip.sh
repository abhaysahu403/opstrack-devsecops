#!/usr/bin/env bash
# generate-zip.sh - Packages the reports/ directory into opstrack-ci-reports.zip
#
# Usage: scripts/generate-zip.sh [reports_dir] [output_zip]

set -euo pipefail

REPORTS_DIR="${1:-reports}"
OUTPUT_ZIP="${2:-$REPORTS_DIR/opstrack-ci-reports.zip}"

if [ ! -d "$REPORTS_DIR" ]; then
  echo "Reports directory '$REPORTS_DIR' does not exist" >&2
  exit 1
fi

rm -f "$OUTPUT_ZIP"

# Zip the contents of reports/ (not the reports/ folder itself), excluding
# any previous zip so it doesn't try to include itself.
( cd "$REPORTS_DIR" && zip -r -q "$(basename "$OUTPUT_ZIP")" . -x "$(basename "$OUTPUT_ZIP")" )

echo "Wrote $OUTPUT_ZIP"
