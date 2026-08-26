#!/usr/bin/env python3
"""
junit_to_html.py - Converts a JUnit XML file into a simple styled HTML page.

Standard library only (xml.etree). Used for the unit test report so the
dashboard has something more useful to link to than a raw XML file.

Usage:
    python3 scripts/junit_to_html.py --input reports/unit-tests/test-results.xml \
        --output reports/unit-tests/index.html
"""
import argparse
import html
import os
import sys
import xml.etree.ElementTree as ET


PAGE_TOP = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Unit Test Report - OpsTrack CI/CD</title>
<style>
  body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; background:#f6f8fa; margin:0; color:#1f2328; }}
  header {{ background:#0d1117; color:#fff; padding:24px 32px; }}
  header a {{ color:#9fc5ff; text-decoration:none; font-size:14px; }}
  header h1 {{ margin:8px 0 0 0; font-size:22px; }}
  main {{ max-width: 1000px; margin: 24px auto; padding: 0 16px 48px 16px; }}
  table {{ width:100%; border-collapse: collapse; background:#fff; border:1px solid #d0d7de; border-radius:8px; overflow:hidden; }}
  th, td {{ text-align:left; padding:10px 14px; border-bottom:1px solid #eaeef2; font-size:13.5px; }}
  th {{ background:#f6f8fa; }}
  tr.pass td.status {{ color:#1a7f37; font-weight:600; }}
  tr.fail td.status {{ color:#cf222e; font-weight:600; }}
  tr.skipped td.status {{ color:#9a6700; font-weight:600; }}
  .summary {{ display:flex; gap:16px; margin-bottom:20px; flex-wrap:wrap; }}
  .stat {{ background:#fff; border:1px solid #d0d7de; border-radius:8px; padding:14px 20px; min-width:120px; }}
  .stat .n {{ font-size:24px; font-weight:700; }}
  .stat .l {{ font-size:12px; color:#57606a; text-transform:uppercase; letter-spacing:.04em; }}
  pre {{ background:#0d1117; color:#e6edf3; padding:10px; border-radius:6px; font-size:12px; overflow-x:auto; margin:6px 0 0 0; }}
</style>
</head>
<body>
<header>
  <a href="index.html">&larr; OpsTrack CI/CD Dashboard</a>
  <h1>Unit Test Report</h1>
</header>
<main>
"""

PAGE_BOTTOM = """
</main>
</body>
</html>
"""


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    if not os.path.isfile(args.input):
        # No JUnit file (e.g. gotestsum wasn't available) - emit a NOT_CONFIGURED page.
        os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
        with open(args.output, "w") as fh:
            fh.write(PAGE_TOP + "<p>No test-results.xml was found.</p>" + PAGE_BOTTOM)
        with open(os.path.join(os.path.dirname(args.output), "status.txt"), "w") as fh:
            fh.write("SKIPPED\n")
        print("No JUnit input found; wrote placeholder report.")
        return 0

    tree = ET.parse(args.input)
    root = tree.getroot()
    suites = [root] if root.tag == "testsuite" else list(root.findall("testsuite"))

    total = failures = errors = skipped = 0
    rows = []
    for suite in suites:
        total += int(suite.attrib.get("tests", 0))
        failures += int(suite.attrib.get("failures", 0))
        errors += int(suite.attrib.get("errors", 0))
        skipped += int(suite.attrib.get("skipped", 0))
        for case in suite.findall("testcase"):
            name = case.attrib.get("classname", "") + " / " + case.attrib.get("name", "")
            time_s = case.attrib.get("time", "0")
            fail_el = case.find("failure")
            err_el = case.find("error")
            skip_el = case.find("skipped")
            if fail_el is not None or err_el is not None:
                status = "fail"
                msg = (fail_el.attrib.get("message") if fail_el is not None else err_el.attrib.get("message", "")) or ""
                detail = (fail_el.text or err_el.text or "") if (fail_el is not None or err_el is not None) else ""
            elif skip_el is not None:
                status = "skipped"
                msg = skip_el.attrib.get("message", "")
                detail = ""
            else:
                status = "pass"
                msg = ""
                detail = ""
            rows.append((status, name, time_s, msg, detail))

    passed = total - failures - errors - skipped
    overall_status = "FAIL" if (failures or errors) else "PASS"

    parts = [PAGE_TOP]
    parts.append('<div class="summary">')
    for label, n in [("Total", total), ("Passed", passed), ("Failed", failures + errors), ("Skipped", skipped)]:
        parts.append(f'<div class="stat"><div class="n">{n}</div><div class="l">{label}</div></div>')
    parts.append("</div>")

    parts.append("<table><thead><tr><th>Status</th><th>Test</th><th>Time (s)</th></tr></thead><tbody>")
    for status, name, time_s, msg, detail in rows:
        parts.append(f'<tr class="{status}"><td class="status">{status.upper()}</td>'
                      f'<td>{html.escape(name)}{("<pre>" + html.escape(msg + chr(10) + detail) + "</pre>") if status == "fail" and (msg or detail) else ""}</td>'
                      f'<td>{html.escape(time_s)}</td></tr>')
    parts.append("</tbody></table>")
    parts.append(PAGE_BOTTOM)

    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
    with open(args.output, "w") as fh:
        fh.write("".join(parts))
    with open(os.path.join(os.path.dirname(args.output), "status.txt"), "w") as fh:
        fh.write(overall_status + "\n")

    print(f"Wrote {args.output} [{overall_status}] total={total} passed={passed} failed={failures + errors} skipped={skipped}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
