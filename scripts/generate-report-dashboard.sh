#!/usr/bin/env bash
# generate-report-dashboard.sh
#
# Builds reports/index.html, the central OpsTrack CI/CD Quality Dashboard,
# by reading a `status.txt` file (PASS | FAIL | SKIPPED | NOT_CONFIGURED)
# out of each reports/<tool>/ subdirectory. Each subdirectory + status.txt
# is written earlier in the pipeline by scripts/render_report.py,
# scripts/junit_to_html.py, or `go tool cover -html`.
#
# Usage:
#   scripts/generate-report-dashboard.sh [reports_dir] [build_status]
#
#   reports_dir   defaults to "reports"
#   build_status  overall PASS/FAIL for the "Build Status" line at the top,
#                 defaults to PASS unless any stage below reports FAIL.

set -euo pipefail

REPORTS_DIR="${1:-reports}"
BUILD_STATUS_OVERRIDE="${2:-}"

# Ordered list of report stages: "folder|Display Name"
STAGES=(
  "unit-tests|Unit Tests"
  "coverage|Coverage"
  "sonar|SonarCloud"
  "trivy-fs|Trivy Filesystem Scan"
  "trivy-image|Trivy Image Scan"
  "snyk|Snyk"
  "gosec|Gosec"
)

mkdir -p "$REPORTS_DIR"

status_of() {
  local dir="$1"
  if [ -f "$REPORTS_DIR/$dir/status.txt" ]; then
    tr -d '[:space:]' < "$REPORTS_DIR/$dir/status.txt"
  else
    echo "NOT_CONFIGURED"
  fi
}

badge_color() {
  case "$1" in
    PASS) echo "#1a7f37" ;;
    FAIL) echo "#cf222e" ;;
    SKIPPED) echo "#9a6700" ;;
    *) echo "#57606a" ;;
  esac
}

badge_label() {
  case "$1" in
    NOT_CONFIGURED) echo "NOT CONFIGURED" ;;
    *) echo "$1" ;;
  esac
}

overall="PASS"
rows=""
for entry in "${STAGES[@]}"; do
  dir="${entry%%|*}"
  label="${entry##*|}"
  st="$(status_of "$dir")"
  color="$(badge_color "$st")"
  blabel="$(badge_label "$st")"
  if [ "$st" = "FAIL" ]; then overall="FAIL"; fi

  link="$dir/index.html"
  if [ ! -f "$REPORTS_DIR/$link" ]; then
    link="#"
  fi

  rows="${rows}
    <tr>
      <td>${label}</td>
      <td><span class=\"badge\" style=\"background:${color}\">${blabel}</span></td>
      <td>$( [ "$link" != "#" ] && echo "<a href=\"${link}\">Open report</a>" || echo "<span class=\"muted\">not generated</span>" )</td>
    </tr>"
done

if [ -n "$BUILD_STATUS_OVERRIDE" ]; then
  overall="$BUILD_STATUS_OVERRIDE"
fi

overall_color="$(badge_color "$overall")"
now="$(date -u +"%Y-%m-%d %H:%M:%S UTC")"

cat > "$REPORTS_DIR/index.html" <<HTML
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>OpsTrack CI/CD Quality Dashboard</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; background:#f6f8fa; margin:0; color:#1f2328; }
  header { background:#0d1117; color:#fff; padding:32px; text-align:center; }
  header h1 { margin:0 0 6px 0; font-size:26px; }
  header p { margin:0; color:#9fa6b2; font-size:13px; }
  main { max-width: 860px; margin: 28px auto; padding: 0 16px 56px 16px; }
  .badge { display:inline-block; padding:4px 14px; border-radius:999px; color:#fff; font-weight:600; font-size:13px; letter-spacing:.03em; }
  .top-status { text-align:center; margin: -18px auto 24px auto; }
  table { width:100%; border-collapse: collapse; background:#fff; border:1px solid #d0d7de; border-radius:8px; overflow:hidden; }
  th, td { text-align:left; padding:12px 16px; border-bottom:1px solid #eaeef2; font-size:14px; }
  th { background:#f6f8fa; font-size:12px; text-transform:uppercase; letter-spacing:.04em; color:#57606a; }
  tr:last-child td { border-bottom:none; }
  .muted { color:#8c959f; }
  .actions { margin-top:24px; display:flex; gap:12px; flex-wrap:wrap; }
  .actions a { background:#0d1117; color:#fff; text-decoration:none; padding:10px 18px; border-radius:6px; font-size:14px; }
  .actions a.secondary { background:#fff; color:#0d1117; border:1px solid #d0d7de; }
  footer { text-align:center; font-size:12px; color:#8c959f; margin-top:32px; }
</style>
</head>
<body>
<header>
  <h1>OpsTrack CI/CD Quality Dashboard</h1>
  <p>Generated ${now}</p>
</header>
<main>
  <div class="top-status">
    <span class="badge" style="background:${overall_color}; font-size:16px; padding:8px 22px;">Build Status: ${overall}</span>
  </div>

  <table>
    <thead><tr><th>Stage</th><th>Status</th><th>Report</th></tr></thead>
    <tbody>${rows}
    </tbody>
  </table>

  <div class="actions">
    <a href="opstrack-ci-reports.zip">Download All Reports (ZIP)</a>
    <a class="secondary" href="https://github.com" target="_blank" rel="noopener">View GitHub Actions Run</a>
  </div>

  <footer>OpsTrack DevSecOps Prototype &mdash; report/security tool statuses reflect actual tool execution in this run; stages marked NOT CONFIGURED require a secret/token that was not present.</footer>
</main>
</body>
</html>
HTML

echo "Dashboard written to $REPORTS_DIR/index.html (overall: $overall)"
