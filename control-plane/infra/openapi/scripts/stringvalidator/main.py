import os
import sys

from openapi_validator import OpenAPIValidator
from utils import load_yaml_file
from const import OPENAPI_FILE_ENV_VAR, EXIT_SUCCESS, EXIT_FAILURE, VALID_NUMBER_OF_ERRORS, valid_regexes


def handle_validation_result(errors):
    if errors:
        for error in errors:
            print(f"- {error}")
        print(f"Found {len(errors)} validation errors")
        if len(errors) > VALID_NUMBER_OF_ERRORS:
            return EXIT_FAILURE
        else:
            # until we fix all existing issues
            return EXIT_SUCCESS

    else:
        return EXIT_SUCCESS

def main():
    openapi_file_path = os.environ.get(OPENAPI_FILE_ENV_VAR)

    if not openapi_file_path:
        print(f"Missing environment variable: {OPENAPI_FILE_ENV_VAR}")
        return EXIT_FAILURE

    openapi_spec = load_yaml_file(openapi_file_path)

    if not openapi_spec:
        print("Failed to load OpenAPI spec.")
        return EXIT_FAILURE

    validator = OpenAPIValidator(openapi_spec, valid_regexes)
    errors = validator.validate_post_put_apis()

    return handle_validation_result(errors)

if __name__ == "__main__":
    sys.exit(main())
