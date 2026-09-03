#!/usr/bin/env python3
"""Validates CODEOWNERS file syntax."""

import re
import sys

OWNER_PATTERN = re.compile(
    r'^(@[\w.-]+(/[\w.-]+)?|[\w.+-]+@[\w.-]+\.[a-zA-Z]+)$'
)


def validate(path: str) -> list[str]:
    errors = []
    with open(path) as f:
        for i, raw in enumerate(f, 1):
            line = raw.rstrip('\n')
            stripped = line.strip()
            if not stripped or stripped.startswith('#'):
                continue
            parts = stripped.split()
            for owner in parts[1:]:
                if not OWNER_PATTERN.match(owner):
                    errors.append(f'Line {i}: invalid owner format: {owner!r}')
    return errors


if __name__ == '__main__':
    path = sys.argv[1] if len(sys.argv) > 1 else 'CODEOWNERS'
    errors = validate(path)
    if errors:
        for e in errors:
            print(e)
        sys.exit(1)
    print('CODEOWNERS validation passed.')