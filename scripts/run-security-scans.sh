#!/usr/bin/env bash
# run-security-scans.sh
#
# Runs the security tools locally (whichever are installed) and writes
# reports into reports/, mirroring the layout CI produces. Useful for
# checking things before opening a PR. Every tool is optional - missing
# tools are skipped with a clear message rather than failing the script.
#
# Requires (all optional, install what you want to test locally):
#   gosec          https://github.com/securego/gosec
#   trivy          https://github.com/aquasecurity/trivy
#   snyk CLI       https://github.com/snyk/cli   (needs `snyk auth` or SNYK_TOKEN)
#   python3        used by scripts/render_report.py

set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

REPORTS_DIR="reports"
mkdir -p "$REPORTS_DIR"

echo "== gosec =="
if command -v gosec >/dev/null 2>&1; then
  mkdir -p "$REPORTS_DIR/gosec"
  gosec -fmt=json -out="$REPORTS_DIR/gosec/gosec.json" ./... > "$REPORTS_DIR/gosec/gosec.txt" 2>&1
  gosec_status=$?
  status="PASS"; [ "$gosec_status" -ne 0 ] && status="FAIL"
  python3 scripts/render_report.py --title "Gosec Security Scan" --tool gosec --status "$status" \
    --summary "See raw output below." --details-file "$REPORTS_DIR/gosec/gosec.txt" \
    --raw-link gosec.json --output "$REPORTS_DIR/gosec/index.html"
else
  echo "gosec not installed - skipping (CI will still run it). Install: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi

echo "== trivy (filesystem) =="
if command -v trivy >/dev/null 2>&1; then
  mkdir -p "$REPORTS_DIR/trivy-fs"
  trivy fs --format json -o "$REPORTS_DIR/trivy-fs/trivy-fs.json" . > "$REPORTS_DIR/trivy-fs/trivy-fs.txt" 2>&1
  status="PASS"
  python3 scripts/render_report.py --title "Trivy Filesystem Scan" --tool trivy-fs --status "$status" \
    --summary "Filesystem/dependency scan complete." --raw-link trivy-fs.json \
    --output "$REPORTS_DIR/trivy-fs/index.html"
else
  echo "trivy not installed - skipping (CI will still run it). Install: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
fi

echo "== snyk =="
if command -v snyk >/dev/null 2>&1 && [ -n "${SNYK_TOKEN:-}" ]; then
  mkdir -p "$REPORTS_DIR/snyk"
  snyk test --json-file-output="$REPORTS_DIR/snyk/snyk.json" > "$REPORTS_DIR/snyk/snyk.txt" 2>&1
  status="PASS"; [ $? -ne 0 ] && status="FAIL"
  python3 scripts/render_report.py --title "Snyk Dependency Scan" --tool snyk --status "$status" \
    --summary "See raw output below." --details-file "$REPORTS_DIR/snyk/snyk.txt" \
    --raw-link snyk.json --output "$REPORTS_DIR/snyk/index.html"
else
  echo "snyk not installed or SNYK_TOKEN not set - skipping locally (CI uses the SNYK_TOKEN secret)."
fi

echo
echo "Done. Open $REPORTS_DIR/ to see whatever reports were generated."
echo "Run scripts/generate-report-dashboard.sh afterwards to (re)build reports/index.html."
