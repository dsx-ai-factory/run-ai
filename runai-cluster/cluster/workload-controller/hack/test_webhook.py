import os
import sys
import requests
import json
import yaml
import argparse
from pprint import pprint
from urllib3.exceptions import InsecureRequestWarning





if __name__ == '__main__':

    # Suppress only the single warning from urllib3 needed.
    requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)


    # Define the allowed operation types
    allowed_operations = ['CREATE', 'UPDATE', 'DELETE']

    # Create the argument parser
    parser = argparse.ArgumentParser(description='Simulate a webhook request')

    # Add the positional argument for the CR JSON file path
    parser.add_argument('cr', metavar='CR_json_file', type=str,
                        help='The path of the CR File for submitting')

    # Add the optional argument for the operation type with a default value of 'CREATE'
    parser.add_argument('-t', '--operation', type=str, choices=allowed_operations,
                        default='CREATE', help='The type of request: CREATE/UPDATE/DELETE (default: CREATE)')
    parser.add_argument('-p', '--old-cr', type=str,
                        default='', help='If operation is UPDATE, the path of the old CR File for submitting,'
                                               'otherwise, it is ignored. in case empty the old CR will be the current '
                                               'CR')
    # Parse the command-line arguments
    args = parser.parse_args()


    # Access the parsed arguments
    cr_path = args.cr
    operation_type = args.operation
    old_cr_path = args.old_cr



    with open(cr_path, "r") as fObject:
        objectStr = fObject.read()
    extension = os.path.splitext(cr_path)[1]

    if extension in ['.yaml', '.yml']:
        object = yaml.load(objectStr, Loader=yaml.FullLoader)
    else:
        object = json.loads(objectStr)
    if old_cr_path != '':
        with open(old_cr_path, "r") as fObject:
            old_objectStr = fObject.read()
        extension = os.path.splitext(old_cr_path)[1]
        if extension in ['.yaml', '.yml']:
            old_object = yaml.load(old_objectStr, Loader=yaml.FullLoader)
        else:
            old_object = json.loads(old_objectStr)
    else:
        old_object = object
    kind = object["kind"]
    name = object["metadata"]["name"]
    namespace = object["metadata"]["namespace"]

    if kind == "InteractivePolicy":
        resource = "interactivepolicy"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-interactivepolicy"
    elif kind == "InteractiveWorkload":
        resource = "interactiveworkload"
        url = "https://localhost:9443/validate-interactive-workload"
    elif kind == "TrainingPolicy":
        resource = "trainingpolicy"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-trainingpolicy"
    elif kind == "TrainingWorkload":
        resource = "trainingworkload"
        url = "https://localhost:9443/validate-training-workload"
    elif kind == "DistributedPolicy":
        resource = "distributedpolicies"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-distributedpolicy"
    elif kind == "DistributedWorkload":
        resource = "distributedworkloads"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-distributedworkload"
    elif kind == "InferencePolicy":
        resource = "inferencepolicy"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-inferencepolicy"
    elif kind == "InferenceWorkload":
        resource = "inferenceworkloads"
        url = "https://localhost:9443/validate-run-ai-v2alpha1-inferenceworkload"
    else:
        print("Usage: unsupported resource kind '%s'" % kind)
        sys.exit(1)

    request = {
      "kind": "AdmissionReview",
      "apiVersion": "admission.k8s.io/v1",
      "request": {
        "uid": "5d660c21-8868-49a5-9d16-9aab8b25363b",
        "kind": {
          "group": "run.ai",
          "version": "v2alpha1",
          "kind": kind
        },
        "resource": {
          "group": "run.ai",
          "version": "v2alpha1",
          "resource": resource
        },
        "requestKind": {
          "group": "run.ai",
          "version": "v2alpha1",
          "kind": kind
        },
        "requestResource": {
          "group": "",
          "version": "v2alpha1",
          "resource": resource
        },
        "name": name,
        "namespace": namespace,
        "operation": operation_type,
        "userInfo": {
          "username": "system:serviceaccount:runai:runai-job-controller",
          "uid": "acf68a42-0f25-4a5b-8474-1eb6dd4cc685",
          "groups": [
            "system:serviceaccounts",
            "system:serviceaccounts:runai",
            "system:authenticated"
          ],
          "extra": {
            "authentication.kubernetes.io/pod-name": [
              "runai-job-controller-5df848c7c-tx8nc"
            ],
            "authentication.kubernetes.io/pod-uid": [
              "81afe7f9-5eee-4945-a30f-4d0769ab23a8"
            ]
          }
        },
        "object": object,
        "oldObject": old_object,
        "dryRun": False,
        "options": {
          "kind": "CreateOptions",
          "apiVersion": "meta.k8s.io/v1"
        }
      }
    }

    payload = json.dumps(request, indent=4)

    headers = {
      'Content-Type': 'application/json'
    }

    print(url)
    print(payload)

    # If you're using self-signed certificates or testing environments, pass verify=False to skip SSL verification
    response = requests.request("POST", url, headers=headers, data=payload)

    print("Response:")
    try:
        print(json.dumps(json.loads(response.text), indent=4))
    except Exception as e:
        print(response.text)
