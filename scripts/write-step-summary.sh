#!/usr/bin/env bash
# write-step-summary.sh
#
# Appends a Markdown summary to $GITHUB_STEP_SUMMARY using the status.txt
# files produced by each report stage. Safe to run outside GitHub Actions
# (prints to stdout instead).
#
# Usage:
#   scripts/write-step-summary.sh [reports_dir] [pages_url] [run_url]

set -euo pipefail

REPORTS_DIR="${1:-reports}"
PAGES_URL="${2:-}"
RUN_URL="${3:-${GITHUB_SERVER_URL:-}/${GITHUB_REPOSITORY:-}/actions/runs/${GITHUB_RUN_ID:-}}"

status_of() {
  local dir="$1"
  if [ -f "$REPORTS_DIR/$dir/status.txt" ]; then
    tr -d '[:space:]' < "$REPORTS_DIR/$dir/status.txt"
  else
    echo "NOT_CONFIGURED"
  fi
}

coverage_pct="n/a"
if [ -f "$REPORTS_DIR/coverage/coverage-percent.txt" ]; then
  coverage_pct="$(cat "$REPORTS_DIR/coverage/coverage-percent.txt")"
fi

UNIT="$(status_of unit-tests)"
GOSEC="$(status_of gosec)"
TRIVY_FS="$(status_of trivy-fs)"
TRIVY_IMAGE="$(status_of trivy-image)"
SNYK="$(status_of snyk)"
SONAR="$(status_of sonar)"
DOCKER_BUILD="${DOCKER_BUILD_STATUS:-PASS}"

overall="PASS"
for s in "$UNIT" "$GOSEC" "$TRIVY_FS" "$TRIVY_IMAGE" "$SNYK" "$SONAR" "$DOCKER_BUILD"; do
  if [ "$s" = "FAIL" ]; then overall="FAIL"; fi
done

{
  echo "# OpsTrack CI/CD Report"
  echo
  echo "## Build"
  echo
  echo "${overall}"
  echo
  echo "## Tests"
  echo
  echo "Unit Tests: ${UNIT}"
  echo "Coverage: ${coverage_pct}%"
  echo
  echo "## Security"
  echo
  echo "Gosec: ${GOSEC}"
  echo "Trivy (filesystem): ${TRIVY_FS}"
  echo "Trivy (image): ${TRIVY_IMAGE}"
  echo "Snyk: ${SNYK}"
  echo
  echo "## Code Quality"
  echo
  echo "SonarCloud: ${SONAR}"
  echo
  echo "## Docker"
  echo
  echo "Image Build: ${DOCKER_BUILD}"
  echo
  echo "## Reports"
  echo
  if [ -n "$PAGES_URL" ]; then
    echo "- [View Dashboard](${PAGES_URL})"
  else
    echo "- View Dashboard: not published on this run (see the reports ZIP artifact instead)"
  fi
  echo "- [Download ZIP / all reports](${RUN_URL})"
  echo "- [Unit Test Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}unit-tests/" || echo "${RUN_URL}"))"
  echo "- [Coverage Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}coverage/" || echo "${RUN_URL}"))"
  echo "- [Trivy Filesystem Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}trivy-fs/" || echo "${RUN_URL}"))"
  echo "- [Trivy Image Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}trivy-image/" || echo "${RUN_URL}"))"
  echo "- [Snyk Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}snyk/" || echo "${RUN_URL}"))"
  echo "- [Gosec Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}gosec/" || echo "${RUN_URL}"))"
  echo "- [SonarCloud Report]($( [ -n "$PAGES_URL" ] && echo "${PAGES_URL}sonar/" || echo "${RUN_URL}"))"
} >> "${GITHUB_STEP_SUMMARY:-/dev/stdout}"

echo "Step summary written (overall: ${overall})"
