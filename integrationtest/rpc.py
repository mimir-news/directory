# 3rd party modules
import requests


def make_request(test_data, env):
    method = test_data["method"]
    if method in ("GET", "DELETE"):
        return _make_request_without_body(test_data, env)
    return _make_request_with_body(test_data, env)


def _make_path(test_data, env):
    path = env["baseUrl"] + test_data["path"]
    for key, val in env.items():
        path_key = "${" + key + "}"
        path = path.replace(path_key, val)
    return path


def _make_request_without_body(test_data, env):
    url = _make_path(test_data, env)
    method = test_data["method"]
    headers = _make_headers(env, test_data["withToken"])
    return requests.request(method, url, headers=headers)


def _make_request_with_body(test_data, env):
    if not "body" in test_data:
        return _make_request_without_body(test_data, env)
    url = _make_path(test_data, env)
    method = test_data["method"]
    headers = _make_headers(env, test_data["withToken"])
    body = test_data["body"]
    return requests.request(method, url, json=body, headers=headers)


def _make_headers(env, with_token=True):
    base_headers = {"Content-Type": "application/json", "X-Client-ID": env["clientId"]}
    if with_token:
        base_headers["Authorization"] = f'Bearer {env["authToken"]}'
    return base_headers
