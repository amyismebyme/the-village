#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

python3 - "$root" <<'PY'
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
forbidden = re.compile(r"(?i)(access[_-]?token|refresh[_-]?token|client[_-]?secret|authorization|password|cookie|credential|dsn|connection[_-]?string)")
telemetry_file = re.compile(r"/(internal/(metrics|observability|telemetry|logger)/|infra/docker/prometheus/)")
definition = re.compile(r"(?i)(NewCounterVec|NewGaugeVec|NewHistogramVec|attribute\.|WithAttributes|labels:|labelnames:)")
violations = []

for path in root.rglob('*'):
    if not path.is_file() or path.suffix.lower() not in {'.go', '.yml', '.yaml'}:
        continue
    if '/.git/' in path.as_posix() or path.name == 'api.exe' or not telemetry_file.search(path.as_posix()):
        continue
    for n, line in enumerate(path.read_text(errors='ignore').splitlines(), 1):
        if forbidden.search(line) and definition.search(line):
            violations.append(f'{path}:{n}: {line.strip()}')

if violations:
    print('\n'.join(violations), file=sys.stderr)
    raise SystemExit(1)

print('Observability security scan passed.')
PY
