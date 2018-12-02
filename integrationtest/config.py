# Standard library
import json
import sys


def read_test_case_content():
    with open("./conf/test_cases.json", "r") as f:
        return json.load(f)


def get_test_cases():
    _content = read_test_case_content()
    return _content["tests"]


def get_env():
    _content = read_test_case_content()
    env = _content["env"]
    env["baseUrl"] = f"http://127.0.0.1:{sys.argv[1]}"
    return env


env = get_env()
TESTS = get_test_cases()
