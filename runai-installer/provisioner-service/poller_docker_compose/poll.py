import requests
import json
import time
import os
import logging
from dataclasses import dataclass, asdict
from typing import Optional

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler()]
)
logger = logging.getLogger(__name__)

BASE_URL = os.environ.get("BASE_URL", "")
DOMAIN = os.environ.get("DOMAIN", "")
WILDCARD_DOMAIN = os.environ.get("WILDCARD_DOMAIN", "")
PROVISIONING_MODE = os.environ.get("PROVISIONING_MODE", "")
USER_EMAIL = os.environ.get("USER_EMAIL", "")
KUBERNETES_DISTRIBUTION = os.environ.get("KUBERNETES_DISTRIBUTION", "")
POLL_INTERVAL = int(os.environ.get("POLL_INTERVAL", "5"))


COMPLETION_PHASES = ["succeeded", "failed"]
FAILED_PHASE = "failed"


@dataclass
class CreateEnvironmentRequest:
    provisioningMode: str
    domain: str
    wildcarddomain: str
    userEmail: str
    kubernetesDistribution: Optional[str] = None


@dataclass
class EnvironmentResponse:
    id: str
    phase: str
    message: Optional[str] = None


def create_cluster() -> EnvironmentResponse:
    url = f"{BASE_URL}/environments"
    request = CreateEnvironmentRequest(
        provisioningMode=PROVISIONING_MODE,
        domain=DOMAIN,
        wildcarddomain=WILDCARD_DOMAIN,
        userEmail=USER_EMAIL,
        kubernetesDistribution=KUBERNETES_DISTRIBUTION or None,
    )
    payload = {k: v for k, v in asdict(request).items() if v is not None}
    
    response = requests.post(url, json=payload)
    logger.info(json.dumps(response.json(), indent=2))
    try:
        response.raise_for_status()
    except requests.exceptions.HTTPError as e:
        logger.error(f"Error creating cluster: {e}")
        logger.error(json.dumps(response.json(), indent=2))
        raise e
    data = response.json()
    return EnvironmentResponse(id=data["id"], phase=data["phase"])

def get_environment(env_id: str) -> EnvironmentResponse:
    url = f"{BASE_URL}/environments/{env_id}"
    response = requests.get(url)
    response.raise_for_status()
    data = response.json()
    return EnvironmentResponse(id=data["id"], phase=data["phase"])

def poll_until_complete(env_id: str) -> str:
    while True:
        result = get_environment(env_id)
        logger.info(json.dumps(asdict(result), indent=2))
        
        if result.phase in COMPLETION_PHASES:
            return result.phase
        
        time.sleep(POLL_INTERVAL)

if __name__ == "__main__":
    import sys
    create_response = create_cluster()
    logger.info(f"Created environment: {create_response.id}")
    logger.info(json.dumps(asdict(create_response), indent=2))
    logger.info("Polling for status...")
    
    final_phase = poll_until_complete(create_response.id)
    if final_phase == FAILED_PHASE:
        sys.exit(1)


