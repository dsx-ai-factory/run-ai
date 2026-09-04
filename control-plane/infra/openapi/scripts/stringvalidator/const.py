EXIT_SUCCESS = 0
EXIT_FAILURE = 1
VALID_NUMBER_OF_ERRORS = 99
OPENAPI_FILE_ENV_VAR = 'SPEC_FILE_SRC'

# List of valid regex patterns with comments for readability
valid_regexes = [
    # Kubernetes Names Regex
    "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$",
    # Runai Names Regex
    "^[a-z][-_a-z0-9]*[a-z0-9]$",
    # Kubernetes Names Regex With Length
    "^[a-z][-a-z0-9]{0,62}$",
    # Free Text Regex
    ".*",
    # Memory Quantity Regex
    "^([0-9]+(\\.[0-9]+)?)(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|K|e|p|t|g|m|k|)$",
    # Email Regex
    "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
    # UUID Regex
    "\\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\\b",
    # Hash Regex
    "^[a-fA-F0-9]{32,128}$",
    # DateTime Regex
    "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):([0-5]\\d):([0-5]\\d)([+-](0\\d|1[0-2]):[0-5]\\d|Z)$",
    # Identity Provider Base URL
    "/^(https?://)?([a-zA-Z0-9.-]+\\.[a-zA-Z]{2,})(:[0-9]{1,5})?(/[^\\s]*)?$|^$/",
    "/^(https?://)?([a-zA-Z0-9.-]+\\.[a-zA-Z]{2,})(:[0-9]{1,5})?(/[^\\s]*)?$|^$/",
    # OIDC Scopes
    "^([a-zA-Z0-9-._:]+)(\\s[a-zA-Z0-9-._:]+)*$",
    # Org Filter
    "^(projectName|departmentName|clusterId)(==).+$",
    # Duration (e.g "2d4.5h", "+1h", "30m200s", "5ms")
    "^(0|[+]?((\\d+(\\.\\d+)?)(ns|us|µs|μs|ms|s|m|h|d|w))+$)",
    # Memory strings (e.g "2Gi", "500Mi", "1.5Ti")
    "^([+]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$",
    # string for k82 labels
    "^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$",
    # Path to access HTTP server
    "^(\\x2F[a-zA-Z0-9\\-_.\\x2F]*)?$"
]
