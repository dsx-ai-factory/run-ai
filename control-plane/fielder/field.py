from dataclasses import dataclass, field
from typing import Set, List, Dict, Optional

from target import Target

@dataclass
class Variant:
    name: Optional[str] = None
    spec: Optional[str] = None
    defaults: Optional[str] = None
    rules: Optional[str] = None
    targets: Set[str] = field(default_factory=set)
    internal: bool = False
    deprecated: bool = False

    def validate(self, field_id: str, known_targets: Dict[str, Target]):
        if not self.targets:
            raise ValueError(f"{field_id}: Each variant must have at least one target.")
        for t in self.targets:
            if t not in known_targets:
                raise ValueError(f"{field_id}: Unknown target '{t}' in field '{field_id}' variant.")

@dataclass
class Field:
    id: str
    name: str
    variants: List[Variant] = field(default_factory=list)

    def validate(self, known_targets: Dict[str, Target]):
        if not self.name:
            raise ValueError(f"Field {self.id}: Field must have a name.")
        if not self.variants:
            raise ValueError(f"Field '{self.id}' must have at least one variant.")

        # make sure each target an name appear in only one variant
        seen_targets: Set[str] = set()
        seen_names: Set[str] = set()

        for variant in self.variants:
            variant.validate(self.id, known_targets)

            if variant.name:
                if variant.name in seen_names:
                    raise ValueError(f"Field '{self.id}' has multiple variants using the same name: '{variant.name}'")
                seen_names.add(variant.name)

            # Check for duplicate target IDs across variants
            overlap = seen_targets.intersection(variant.targets)
            if overlap:
                raise ValueError(
                    f"Field '{self.id}' has multiple variants using the same target(s): {', '.join(overlap)}"
                )
            seen_targets.update(variant.targets)
