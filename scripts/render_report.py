#!/usr/bin/env python3
"""
render_report.py - Renders a single, self-contained HTML report page for a
CI/security tool stage (unit tests, coverage, gosec, trivy, snyk, sonar...).

This is deliberately dependency-free (Python 3 standard library only) so it
runs on any GitHub Actions runner without an extra `pip install` step.

Usage:
    python3 scripts/render_report.py \
        --title "Gosec Security Scan" \
        --tool gosec \
        --status PASS \
        --summary "0 issues found across 42 files" \
        --details-file reports/gosec/gosec.txt \
        --output reports/gosec/index.html \
        --raw-link gosec.json

Status must be one of: PASS, FAIL, SKIPPED, NOT_CONFIGURED
"""
import argparse
import datetime
import html
import os
import sys

STATUS_COLORS = {
    "PASS": "#1a7f37",
    "FAIL": "#cf222e",
    "SKIPPED": "#9a6700",
    "NOT_CONFIGURED": "#57606a",
}

STATUS_LABELS = {
    "PASS": "PASS",
    "FAIL": "FAIL",
    "SKIPPED": "SKIPPED",
    "NOT_CONFIGURED": "NOT CONFIGURED",
}

PAGE_TEMPLATE = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{title} - OpsTrack CI/CD</title>
<style>
  :root {{ color-scheme: light; }}
  body {{
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    margin: 0; padding: 0; background: #f6f8fa; color: #1f2328;
  }}
  header {{
    background: #0d1117; color: #fff; padding: 24px 32px;
  }}
  header a {{ color: #9fc5ff; text-decoration: none; font-size: 14px; }}
  header h1 {{ margin: 8px 0 0 0; font-size: 22px; }}
  main {{ max-width: 900px; margin: 24px auto; padding: 0 16px 48px 16px; }}
  .badge {{
    display: inline-block; padding: 4px 14px; border-radius: 999px;
    color: #fff; font-weight: 600; font-size: 13px; letter-spacing: 0.03em;
  }}
  .card {{
    background: #fff; border: 1px solid #d0d7de; border-radius: 8px;
    padding: 20px 24px; margin-top: 16px;
  }}
  .meta {{ color: #57606a; font-size: 13px; margin-top: 4px; }}
  pre {{
    background: #0d1117; color: #e6edf3; padding: 16px; border-radius: 6px;
    overflow-x: auto; font-size: 12.5px; line-height: 1.5;
    max-height: 520px;
  }}
  .footer {{ margin-top: 32px; font-size: 12px; color: #8c959f; text-align: center; }}
  a.raw {{ font-size: 13px; }}
</style>
</head>
<body>
<header>
  <a href="index.html">&larr; OpsTrack CI/CD Dashboard</a>
  <h1>{title}</h1>
</header>
<main>
  <div class="card">
    <span class="badge" style="background:{color}">{status_label}</span>
    <p class="meta">Tool: <strong>{tool}</strong> &nbsp;&middot;&nbsp; Generated: {timestamp} UTC</p>
    <p>{summary}</p>
    {raw_link_html}
  </div>
  {details_html}
  <div class="footer">OpsTrack DevSecOps Prototype &mdash; generated report, not a substitute for reviewing the tool's own output.</div>
</main>
</body>
</html>
"""

DETAILS_TEMPLATE = """<div class="card">
    <h3 style="margin-top:0">Details</h3>
    <pre>{details}</pre>
  </div>"""


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--title", required=True)
    parser.add_argument("--tool", required=True)
    parser.add_argument("--status", required=True, choices=list(STATUS_COLORS))
    parser.add_argument("--summary", default="")
    parser.add_argument("--details-file", default=None, help="Plain-text file whose content is shown verbatim")
    parser.add_argument("--details-max-chars", type=int, default=20000)
    parser.add_argument("--raw-link", default=None, help="Filename (relative to the report dir) of the raw machine-readable output")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    details_html = ""
    if args.details_file and os.path.isfile(args.details_file):
        with open(args.details_file, "r", errors="replace") as fh:
            content = fh.read()
        if len(content) > args.details_max_chars:
            content = content[: args.details_max_chars] + "\n... [truncated] ..."
        if content.strip():
            details_html = DETAILS_TEMPLATE.format(details=html.escape(content))

    raw_link_html = ""
    if args.raw_link:
        raw_link_html = f'<p class="raw"><a class="raw" href="{html.escape(args.raw_link)}">Download raw machine-readable output ({html.escape(args.raw_link)})</a></p>'

    page = PAGE_TEMPLATE.format(
        title=html.escape(args.title),
        tool=html.escape(args.tool),
        status_label=STATUS_LABELS[args.status],
        color=STATUS_COLORS[args.status],
        summary=html.escape(args.summary) if args.summary else "No summary provided.",
        timestamp=datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d %H:%M:%S"),
        details_html=details_html,
        raw_link_html=raw_link_html,
    )

    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
    with open(args.output, "w") as fh:
        fh.write(page)

    # Also drop a status.txt next to the report so the dashboard generator
    # and the GitHub Actions summary script can read it without re-parsing HTML.
    status_path = os.path.join(os.path.dirname(args.output), "status.txt")
    with open(status_path, "w") as fh:
        fh.write(args.status + "\n")

    print(f"Wrote {args.output} [{args.status}]")
    return 0


if __name__ == "__main__":
    sys.exit(main())
