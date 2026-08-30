import argparse
import os
from html.parser import HTMLParser

# ANSI escape sequences for terminal colors
COLOR_GREEN = "\033[92m"
COLOR_RED = "\033[91m"
COLOR_RESET = "\033[0m"

class TestReportParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_test_row = False
        self.current_test_name = ""
        self.current_duration = ""
        self.current_result = ""
        self.tests = []

    def handle_starttag(self, tag, attrs):
        if tag == 'tr':
            self.in_test_row = True
        elif tag == 'td' and self.in_test_row:
            # Check if we're parsing a test case
            if not self.current_test_name:
                self.current_test_name = ""
            elif not self.current_duration:
                self.current_duration = ""
            elif not self.current_result:
                self.current_result = ""

    def handle_endtag(self, tag):
        if tag == 'tr' and self.in_test_row:
            self.tests.append((self.current_test_name, self.current_duration, self.current_result))
            self.in_test_row = False
            self.current_test_name = ""
            self.current_duration = ""
            self.current_result = ""

    def handle_data(self, data):
        data = data.strip()
        if self.in_test_row:
            if not self.current_test_name:
                self.current_test_name = data
            elif not self.current_duration:
                self.current_duration = data
            elif not self.current_result:
                self.current_result = data

def parse_html_report(file_path):
    with open(file_path, 'r', encoding='utf-8') as file:
        parser = TestReportParser()
        parser.feed(file.read())
        for test in parser.tests:
            test_name, duration, result = test
            if "pass" in result.lower():
                print(f"{COLOR_GREEN}Test: {test_name}, Duration: {duration}, Result: {result}{COLOR_RESET}")
            elif "fail" in result.lower():
                print(f"{COLOR_RED}Test: {test_name}, Duration: {duration}, Result: {result}{COLOR_RESET}")
            else:
                print(f"Test: {test_name}, Duration: {duration}, Result: {result}")

def parse_reports_in_directory(directory):
    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('.html'):
                file_path = os.path.join(root, file)
                print(f"Parsing file: {file_path}")
                parse_html_report(file_path)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Parse HTML test report files.')
    parser.add_argument('report_dir', type=str, help='Directory containing HTML report files')
    args = parser.parse_args()

    parse_reports_in_directory(args.report_dir)
