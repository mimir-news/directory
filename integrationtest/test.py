# Standard library
import sys
from datetime import datetime
from typing import Dict

# 3rd party modules
import requests

# Internal modules
import rpc
from config import env, TESTS


def _run_test(test_case: Dict, env: Dict) -> None:
    resp = rpc.make_request(test_case["request"], env)
    _assert_status(resp, test_case)
    if test_case["positive"]:
        _update_env(resp, test_case, env)


def _assert_status(resp: requests.Response, test_case: Dict) -> None:
    expected_status = test_case["response"]["status"]
    actual_status = resp.status_code
    if resp.status_code != expected_status:
        test_name = test_case["name"]
        print(f"FAILED: {test_name} expected={expected_status} got={actual_status}\n")
        sys.exit(1)


def _describe_test(test_no: int, test_case: Dict) -> None:
    name = test_case["name"]
    print(f"Test {test_no}: {name}")


def _update_env(resp: requests.Response, test_case: Dict, env: Dict) -> None:
    updates = test_case["setEnv"] if "setEnv" in test_case else []
    if not updates:
        return
    body = resp.json()
    for update in updates:
        env[update["envKey"]] = body[update["responseKey"]]


def _report_elapsed_time(start_time: datetime) -> None:
    elapsed_time = (datetime.now() - start_time).microseconds / 1e6
    print(f"Test took: {elapsed_time} s.")


def main() -> None:
    start_time = datetime.now()
    for i, TEST in enumerate(TESTS):
        _describe_test(i, TEST)
        _run_test(TEST, env)
        print("OK")

    print("\nGreat success!")
    _report_elapsed_time(start_time)


if __name__ == "__main__":
    main()

