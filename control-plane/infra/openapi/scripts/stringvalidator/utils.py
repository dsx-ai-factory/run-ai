import yaml
import re

def load_yaml_file(file_path):
    try:
        with open(file_path, 'r') as file:
            return yaml.safe_load(file)
    except FileNotFoundError:
        print(f"File {file_path} not found.")
        return None
    except yaml.YAMLError as e:
        print(f"Failed to parse YAML: {e}")
        return None

def load_valid_regexes(file_path):
    try:
        with open(file_path, 'r') as file:
            regex_file_content = file.read()
            pattern = r"^(\w+)=\"(.*?)\"$"
            valid_regexes = []
            for line in regex_file_content.splitlines():
                match = re.match(pattern, line)
                if match:
                    key, regex = match.groups()
                    valid_regexes.append(regex)
            return valid_regexes
    except FileNotFoundError:
        print(f"File {file_path} not found.")
        return []
