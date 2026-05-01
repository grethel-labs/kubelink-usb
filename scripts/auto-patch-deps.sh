#!/usr/bin/env bash
#
# Apply patch-level Go module updates and emit a human-readable summary.
#
# Output contract (consumed by the GitHub Actions job):
#   - On no-op: prints exactly  "No patch-level updates available."
#   - On change: prints summary, then a "---SUMMARY---" marker, then
#                a list of "- <module>: <old> -> <new>" lines.
set -euo pipefail

snapshot() {
  go list -m -mod=mod -f '{{if not .Indirect}}{{if not .Main}}{{.Path}}{{"\t"}}{{.Version}}{{end}}{{end}}' all 2>/dev/null \
    | awk 'NF==2' \
    | sort
}

before=$(snapshot)
go get -u=patch ./... >/dev/null 2>&1
go mod tidy >/dev/null 2>&1
after=$(snapshot)

changes=$(BEFORE="$before" AFTER="$after" python3 <<'PY'
import os
before = os.environ["BEFORE"]
after  = os.environ["AFTER"]
b = {l.split("\t",1)[0]: l.split("\t",1)[1] for l in before.splitlines() if "\t" in l}
a = {l.split("\t",1)[0]: l.split("\t",1)[1] for l in after.splitlines()  if "\t" in l}
out = []
for path, new in sorted(a.items()):
    old = b.get(path)
    if old is None or old == new:
        continue
    out.append(f"- {path}: {old} -> {new}")
print("\n".join(out))
PY
)

if [ -z "$changes" ]; then
  echo "No patch-level updates available."
  exit 0
fi

echo "Applied patch-level updates:"
echo "$changes"
echo "---SUMMARY---"
echo "$changes"
