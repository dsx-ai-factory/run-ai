import yaml
import argparse
import sys

def main():
    # Parse command-line arguments
    parser = argparse.ArgumentParser(description="Validate OpenAPI documentation for test coverage.")
    parser.add_argument(
        '--fail-on-missing-tests',
        action='store_true',
        help="Exit with a non-zero status code if any API is missing test coverage."
    )
    parser.add_argument(
        'file',
        help="Path to the OpenAPI YAML file."
    )
    args = parser.parse_args()

    # Load the YAML file
    with open(args.file, 'r') as file:
        openapi_data = yaml.safe_load(file)

    # Initialize a dictionary to store APIs with issues by service
    service_issues = {}

    # Iterate over all paths
    for path, path_data in openapi_data.get('paths', {}).items():
        # Check if the path itself is marked as internal
        if path_data.get('x-internal', False):
            continue

        for method, method_data in path_data.items():
            if method in ['get', 'post', 'put', 'delete', 'patch']:  # Check HTTP methods


                # Check if x-test-coverage is missing or empty
                x_test_coverage = method_data.get('x-test-coverage')
                service = method_data.get('x-service') or method_data.get('x-runai-team', 'Unknown Service')

                # Normalize service name to lowercase and prefix with @
                service = f"@{service.lower()}"

                # Check if x-test-coverage is missing or empty
                if not x_test_coverage:
                    if service not in service_issues:
                        service_issues[service] = []
                    service_issues[service].append({
                        'path': path,
                        'method': method.upper(),
                        'operationId': method_data.get('operationId', 'Unknown Operation')
                    })

    # Output the results
    if service_issues:
        print("APIs without proper test coverage:")
        for service, issues in service_issues.items():
            print(f"\nService: {service}")
            for issue in issues:
                print(f"  - Path: {issue['path']} | Method: {issue['method']} | OperationId: {issue['operationId']}")

        # Exit with a failure status code if the flag is set
        if args.fail_on_missing_tests:
            sys.exit(1)
    else:
        print("")

if __name__ == "__main__":
    main()
