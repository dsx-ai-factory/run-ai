import argparse
import yaml
import sys

from config import Config  # your @dataclass module
from target import WorkingMode

BASE_DIR = ".."

def main():
    parser = argparse.ArgumentParser(description="Build or validate workload fields")
    parser.add_argument("--validate", action="store_true", help="Run in validation mode (do not overwrite target files)")
    args = parser.parse_args()

    working_mode = WorkingMode.VALIDATE if args.validate else WorkingMode.BUILD

    with open("config.yaml", "r") as f:
        config_raw = yaml.safe_load(f)

    config = Config.from_dict(config_raw, BASE_DIR)

    has_differences = False

    for target in config.targets.values():
        diff_found = target.build_file(config.fields, BASE_DIR, working_mode)
        if working_mode == WorkingMode.VALIDATE and diff_found:
            has_differences |= diff_found

    if working_mode == WorkingMode.VALIDATE:
        if has_differences:
            print("\n❌ Validation failed: Some target files are out of sync.", file=sys.stderr)
            sys.exit(1)
        else:
            print("\n✅ All targets are ok"
                  "`.", file=sys.stderr)
    else:
        print("\n✅ All targets build.", file=sys.stderr)
        config.export_excel("config.xlsx")
        print("\n✅ Excel config.xslx created.", file=sys.stderr)
