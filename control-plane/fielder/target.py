import sys
import difflib
from dataclasses import dataclass
from typing import Dict, Optional
from pathlib import Path
from enum import Enum
import os
import yaml

TARGET_TYPES = {"spec", "defaults", "rules"}
WORKLOAD_SPEC_BASE = Path("infra/openapi/docs/schemas/workloads_spec")

class WorkingMode(Enum):
    VALIDATE = "validate"
    BUILD = "build"

@dataclass
class Target:
    id: str
    type: str
    path: str
    comment: str
    skeleton: Optional[str] = None
    name_change: Optional[Dict[str, str]] = None
    nullable: bool = True
    extract_fields: Optional[Dict[str, Dict[str, str]]] = None

    def validate(self, base_dir):
        if None == self.id:
            raise ValueError(f"Mandatory name is missing")

        if None == self.path:
            raise ValueError(f"{self.id}: Mandatory path is not provided: {self.path}")

        if self.type not in TARGET_TYPES:
            raise ValueError(f"{self.id}: Target has invalid type: {self.type}")

        full_path = os.path.join(base_dir, self.path)
        if not full_path.endswith(".yaml"):
            raise ValueError(f"{self.id}: Target path must end with .yaml: {full_path}")

        dir_path = os.path.dirname(full_path)
        if not os.path.isdir(dir_path):
            raise FileNotFoundError(f"{self.id}: Directory does not exist: {dir_path}")

        if self.skeleton:
            skeleton_path = os.path.join("./skeletons", self.skeleton)
            if not os.path.exists(skeleton_path):
                raise FileNotFoundError(f"{self.id}: Skeleton not found: {skeleton_path}")

    def build_file(self, fields: Dict[str, "Field"], base_dir: str, working_mode: WorkingMode) -> bool:
        # Generate the fields section
        fields_yaml = self.build_fields(fields)

        if self.nullable is not None and not self.nullable:
            del fields_yaml['nullable']

        if not self.skeleton:
            final_yaml = fields_yaml
        else:
            skeleton_path = os.path.join("skeletons", self.skeleton)
            with open(skeleton_path, "r") as f:
                skeleton_yaml = yaml.safe_load(f)

            final_yaml = replace_marker(skeleton_yaml, fields_yaml)

        yaml_text = yaml.dump(final_yaml, sort_keys=True)

        # Write the YAML to the output path. in validate write to temp file and compare to
        # the real path
        real_path = os.path.join(base_dir, self.path)
        temp_path = real_path

        if working_mode == WorkingMode.VALIDATE:
            temp_path += ".tmp"

        with open(temp_path, "w") as f:
            f.write("###########################################################\n")
            f.write("# This file is auto-generated do not edit it!!!\n")
            f.write("# Run 'make rebuild' in control-plane/fielder to rebuild it\n")
            f.write("###########################################################\n")
            if self.comment:
                f.write(f"# {self.comment}\n")
            f.write(yaml_text)

        if working_mode == WorkingMode.VALIDATE:
            with open(real_path) as f1, open(temp_path) as f2:
                real = f1.readlines()
                temp = f2.readlines()

            if real != temp:
                print(f"\n[DIFF] {self.id} — {self.path}", file=sys.stderr)
                for line in difflib.unified_diff(real, temp, fromfile="existing", tofile="generated"):
                    print(line, end="", file=sys.stderr)

            os.remove(temp_path)

            return real != temp

        return False

    def build_fields(self, fields: Dict[str, "Field"]) -> dict:
        root = {
            "type": "object",
            "nullable": True,
            "properties": {}
        }

        # Track which extracted fields we've already created refs for
        extracted_refs_created = {}

        for field in fields.values():
            for variant in field.variants:
                if not self.id in variant.targets:
                    continue

                # get the field path that we need to add as $ref
                field_path = getattr(variant, self.type, None)
                if not field_path:
                    raise ValueError(f"Field '{field.id}' variant missing '{self.type}' for target '{self.id}'")

                field_path_rel = compute_ref_path(field_path, self.path)

                path_parts = field.name.split(".")

                if self.name_change:
                    path_parts = [
                        mapped for part in path_parts
                        if (mapped := self.name_change.get(part, part)) != {}
                    ]

                # Generic field extraction: Check if this field should be extracted to a separate file
                if self.extract_fields and len(path_parts) > 0:
                    first_part = path_parts[0]
                    if first_part in self.extract_fields:
                        if first_part not in extracted_refs_created:
                            # Get the ref file from config
                            ref_file = self.extract_fields[first_part].get("refFile")
                            if not ref_file:
                                raise ValueError(f"Target '{self.id}': extractFields.{first_part} missing 'refFile'")
                            
                            # Calculate relative path from target file to the ref file
                            # ref_file is relative to WORKLOAD_SPEC_BASE
                            ref_path_rel = compute_ref_path(ref_file, self.path)
                            root["properties"][first_part] = {
                                "$ref": ref_path_rel
                            }
                            extracted_refs_created[first_part] = True
                        # Skip processing nested fields - we're using a reference instead
                        continue

                current = root["properties"]

                for i, part in enumerate(path_parts):
                    is_last = (i == len(path_parts) - 1)

                    if is_last:
                        if variant.internal:
                            current[part] = {
                                "allOf": [
                                    {
                                        "$ref": field_path_rel
                                    }
                                ],
                                "x-internal": True
                            }
                        elif variant.deprecated:
                            current[part] = {
                                "allOf": [
                                    {
                                        "$ref": field_path_rel
                                    }
                                ],
                                "deprecated": True
                            }
                        else:
                            current[part] = {
                                "$ref": field_path_rel
                            }
                    else:
                        if part not in current:
                            current[part] = {
                                "type": "object",
                                "nullable": True,
                                "properties": {}
                            }

                        current = current[part]["properties"]

        return root

# given two paths. return the relative path from one to another
# for example if field_path is a/b/field.yaml and target path is a/c/target.yaml
# then the computed ref path is ../b/field.path
def compute_ref_path(field_path: str, target_path: str) -> str:
    absolute_ref = (WORKLOAD_SPEC_BASE / field_path).resolve()

    target_dir = Path(target_path).parent.resolve()

    relative = os.path.relpath(absolute_ref, target_dir)
    if relative[0] != ".":
        return "./" + relative

    return relative

def replace_marker(obj, replacement):
    if isinstance(obj, list):
        return [replace_marker(item, replacement) for item in obj]

    if isinstance(obj, dict):
        if obj == {"replace": "me"}:
            return replacement  # <- Replace here

        # Recurse into dict values
        return {
            key: replace_marker(value, replacement)
            for key, value in obj.items()
        }

    return obj  # return as-is for scalars
