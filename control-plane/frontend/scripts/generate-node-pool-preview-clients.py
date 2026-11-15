"""Generate only the node-pool contracts needed by the feature branch.

Run after fetching PR head #4926. The merged cluster capabilities and labels
are generated from the current repository spec, never from an older PR.
Remove this simulation-only bridge when main includes the simulation contract.
Uses the repository's normal OpenAPI generator/templates and postprocessing.
"""
import os
import pathlib
import re
import subprocess
import shutil
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[3]
INPUTS = {
    "org-unit-service": "7310006bb51cd60955f42d72fe95ce0585b166f1",
}


def schema(text, name):
    return re.search(r"^    " + name + r":\n.*?(?=^    \w|\Z)", text, re.M | re.S).group()


for service, revision in INPUTS.items():
    relative = f"control-plane/{service}/swagger/_build/swagger.yaml"
    old = (ROOT / relative).read_text()
    new = subprocess.check_output(["git", "show", f"{revision}:{relative}"], cwd=ROOT, text=True)
    # Already integrated upstream: generate the canonical spec unchanged.
    if "    FairShareSimulationRequest:" not in old:
        names = re.findall(r"^    (FairShareSimulation\w+):", new, re.M)
        old = old.replace("  schemas:\n", "  schemas:\n" + "".join(schema(new, name) for name in names), 1)
        route = re.search(r"^  /api/v1/org-unit/node-pools/\{nodepoolId\}/fair-share-simulation:\n.*?(?=^  /)", new, re.M | re.S).group()
        old = old.replace("paths:\n", "paths:\n" + route, 1)
    temporary = ROOT / f"control-plane/{service}/swagger/_build/step2-client.yaml"
    temporary.write_text(old)
    directory = ROOT / f"control-plane/{service}"
    output = ROOT / f"control-plane/frontend/src/swagger-models/{service}-client"
    backup = pathlib.Path(tempfile.mkdtemp(prefix="nodepool-client-backup-")) / "client"
    shutil.copytree(output, backup)
    try:
        subprocess.run(["make", "openapi-generate-code-client-js", "OPENAPI_GENERATOR_CLIENT_SWAGGER_PATH=swagger/_build/step2-client.yaml"], cwd=directory, env={**os.environ, "PWD": str(directory)}, check=True)
    except Exception:
        # The Make target removes its output before starting Docker. Restore the
        # previous generated client if Docker or generation is unavailable.
        shutil.rmtree(output, ignore_errors=True)
        shutil.copytree(backup, output)
        raise
    finally:
        temporary.unlink()
        shutil.rmtree(backup.parent)
