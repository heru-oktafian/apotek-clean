#!/usr/bin/env python3
import json
import os
import random
import string
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path
from urllib.error import HTTPError, URLError

BASE_URL = os.getenv("BASE_URL", "http://127.0.0.1:9002").rstrip("/")
USERNAME = os.getenv("API_USERNAME", "vita_fauzi")
PASSWORD = os.getenv("API_PASSWORD", "Sigala1102")
BRANCH_ID = os.getenv("API_BRANCH_ID", "")
PURCHASE_ID = os.getenv("SMOKE_PURCHASE_ID", "PUR49589903CJ0G")
SALE_ID = os.getenv("SMOKE_SALE_ID", "SAL466951VV0DS5")
MONTH = os.getenv("SMOKE_MONTH", "2026-02")
EMPTY_MONTH = os.getenv("SMOKE_EMPTY_MONTH", "2099-12")
SMOKE_DATE = os.getenv("SMOKE_DATE", "2026-06-12")
ENABLE_MUTATION = os.getenv("SMOKE_ENABLE_MUTATION", "1") != "0"
ENABLE_DEEP_STOCK = os.getenv("SMOKE_ENABLE_DEEP_STOCK", "1") != "0"
STOCK_TEST_PRODUCT_ID = os.getenv("SMOKE_STOCK_PRODUCT_ID", "PRD028055HKY6YL")
STOCK_TEST_EXPIRED_DATE = os.getenv("SMOKE_STOCK_EXPIRED_DATE", "2027-12-31")
STARTUP_RETRIES = int(os.getenv("SMOKE_STARTUP_RETRIES", "12"))
STARTUP_DELAY_SECONDS = float(os.getenv("SMOKE_STARTUP_DELAY_SECONDS", "1.5"))
READ_RETRIES = int(os.getenv("SMOKE_READ_RETRIES", "4"))
READ_RETRY_DELAY_SECONDS = float(os.getenv("SMOKE_READ_RETRY_DELAY_SECONDS", "0.8"))
MAX_REQUESTS_PER_MINUTE = int(os.getenv("SMOKE_MAX_REQUESTS_PER_MINUTE", "72"))
REQUEST_WINDOW_SECONDS = float(os.getenv("SMOKE_REQUEST_WINDOW_SECONDS", "60"))
CASES_PATH = Path(os.getenv("SMOKE_CASES_PATH", Path(__file__).with_name("regression_cases.json")))
REQUEST_TIMESTAMPS = []


def rnd(n=5):
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=n))


def throttle_request_budget():
    if MAX_REQUESTS_PER_MINUTE <= 0:
        return
    while True:
        now = time.monotonic()
        while REQUEST_TIMESTAMPS and now - REQUEST_TIMESTAMPS[0] >= REQUEST_WINDOW_SECONDS:
            REQUEST_TIMESTAMPS.pop(0)
        if len(REQUEST_TIMESTAMPS) < MAX_REQUESTS_PER_MINUTE:
            REQUEST_TIMESTAMPS.append(now)
            return
        sleep_for = max(REQUEST_WINDOW_SECONDS - (now - REQUEST_TIMESTAMPS[0]) + 0.05, 0.1)
        print(f"[info] request budget penuh, menunggu {sleep_for:.1f}s agar tidak kena rate limit")
        time.sleep(sleep_for)


def request(method, path, token=None, json_payload=None):
    throttle_request_budget()
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    data = None
    if json_payload is not None:
        data = json.dumps(json_payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(BASE_URL + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = resp.read()
            return resp.status, dict(resp.headers), body
    except HTTPError as e:
        return e.code, dict(e.headers), e.read()
    except URLError as e:
        return 0, {}, str(e).encode("utf-8")


def request_with_retry(method, path, token=None, json_payload=None, retries=READ_RETRIES, delay_seconds=READ_RETRY_DELAY_SECONDS):
    last = (0, {}, b"")
    for attempt in range(1, retries + 1):
        status, headers, body = request(method, path, token=token, json_payload=json_payload)
        last = (status, headers, body)
        if status not in (0, 429):
            return last
        if attempt < retries:
            time.sleep(delay_seconds)
    return last


def body_as_json(body):
    try:
        return json.loads(body.decode("utf-8"))
    except Exception:
        return None


def expect_status(name, status, expected):
    ok = status == expected
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: status={status}, expected={expected}")
    return ok


def expect_content_type(name, headers, expected_contains):
    ct = headers.get("Content-Type", "")
    ok = expected_contains in ct
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: content-type={ct!r}, expected~={expected_contains!r}")
    return ok


def expect_data_is_list(name, payload):
    data = payload.get("data") if isinstance(payload, dict) else None
    ok = isinstance(data, list)
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: data_type={type(data).__name__}")
    return ok


def expect_equal(name, actual, expected):
    ok = actual == expected
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: actual={actual}, expected={expected}")
    return ok


def load_cases():
    with CASES_PATH.open("r", encoding="utf-8") as f:
        return json.load(f)


def case_context():
    return {
        "MONTH": MONTH,
        "EMPTY_MONTH": EMPTY_MONTH,
        "PURCHASE_ID": PURCHASE_ID,
        "SALE_ID": SALE_ID,
        "PURCHASE_ID_ENCODED": urllib.parse.quote(PURCHASE_ID),
        "SALE_ID_ENCODED": urllib.parse.quote(SALE_ID),
    }


def render_path(template, ctx):
    return template.format(**ctx)


def run_case_group(branch_token, cases, group_name):
    failures = 0
    ctx = case_context()
    print(f"[info] menjalankan group: {group_name}")
    for case in cases:
        name = case["name"]
        path = render_path(case["path"], ctx)
        expected_status = int(case.get("expected_status", 200))
        expected_content_type = case.get("expected_content_type")
        expect_list = bool(case.get("expect_data_list", False))

        status, headers, body = request_with_retry("GET", path, token=branch_token)
        failures += 0 if expect_status(name, status, expected_status) else 1
        if status == expected_status and expected_content_type:
            failures += 0 if expect_content_type(name, headers, expected_content_type) else 1
        if status == expected_status and expect_list:
            failures += 0 if expect_data_is_list(name, body_as_json(body) or {}) else 1

    return failures


def login_with_retry():
    last = (0, {}, b"")
    for attempt in range(1, STARTUP_RETRIES + 1):
        status, headers, body = request("POST", "/api/login", json_payload={"username": USERNAME, "password": PASSWORD})
        last = (status, headers, body)
        if status == 200:
            if attempt > 1:
                print(f"[info] server siap pada percobaan ke-{attempt}")
            return last
        if status not in (0, 429):
            return last
        time.sleep(STARTUP_DELAY_SECONDS)
    return last


def run_lightweight_mutation_cases(branch_token):
    failures = 0

    expense_payload = {
        "expense_date": SMOKE_DATE,
        "description": f"Smoke expense {rnd()}",
        "total_expense": 12345,
        "payment": "paid_by_cash",
    }
    status, headers, body = request("POST", "/api/expenses", token=branch_token, json_payload=expense_payload)
    failures += 0 if expect_status("expense_create", status, 200) else 1
    expense_data = body_as_json(body) or {}
    expense_id = ((expense_data.get("data") or {}).get("id") if isinstance(expense_data, dict) else None)
    if not expense_id:
        print("[FAIL] expense_id kosong")
        return failures + 1

    expense_update_payload = dict(expense_payload)
    expense_update_payload["description"] = expense_payload["description"] + " updated"
    expense_update_payload["total_expense"] = 23456
    status, headers, body = request("PUT", f"/api/expenses/{expense_id}", token=branch_token, json_payload=expense_update_payload)
    failures += 0 if expect_status("expense_update", status, 200) else 1

    status, headers, body = request("DELETE", f"/api/expenses/{expense_id}", token=branch_token)
    failures += 0 if expect_status("expense_delete", status, 200) else 1

    income_payload = {
        "income_date": SMOKE_DATE,
        "description": f"Smoke income {rnd()}",
        "total_income": 54321,
        "payment": "paid_by_cash",
    }
    status, headers, body = request("POST", "/api/another-incomes", token=branch_token, json_payload=income_payload)
    failures += 0 if expect_status("income_create", status, 200) else 1
    income_data = body_as_json(body) or {}
    income_id = ((income_data.get("data") or {}).get("id") if isinstance(income_data, dict) else None)
    if not income_id:
        print("[FAIL] income_id kosong")
        return failures + 1

    income_update_payload = dict(income_payload)
    income_update_payload["description"] = income_payload["description"] + " updated"
    income_update_payload["total_income"] = 65432
    status, headers, body = request("PUT", f"/api/another-incomes/{income_id}", token=branch_token, json_payload=income_update_payload)
    failures += 0 if expect_status("income_update", status, 200) else 1

    status, headers, body = request("DELETE", f"/api/another-incomes/{income_id}", token=branch_token)
    failures += 0 if expect_status("income_delete", status, 200) else 1

    return failures


def product_payload_from_response(payload):
    data = payload.get("data") if isinstance(payload, dict) else None
    if isinstance(data, list):
        return data[0] if data else {}
    if isinstance(data, dict):
        return data
    return {}


def as_int(value, default=0):
    try:
        return int(value)
    except Exception:
        return default


def get_product_detail(branch_token, product_id):
    status, headers, body = request_with_retry("GET", f"/api/products/{product_id}", token=branch_token)
    payload = body_as_json(body) or {}
    return status, product_payload_from_response(payload)


def run_deep_stock_case(branch_token):
    failures = 0
    product_id = STOCK_TEST_PRODUCT_ID

    status, product_before = get_product_detail(branch_token, product_id)
    failures += 0 if expect_status("stock_product_detail_before", status, 200) else 1
    if status != 200:
        return failures

    stock_before = as_int(product_before.get("stock"), 0)
    unit_id = product_before.get("unit_id") or ((product_before.get("unit") or {}).get("id") if isinstance(product_before.get("unit"), dict) else None)
    price = as_int(product_before.get("purchase_price") or product_before.get("price"), 1000)
    qty = 1
    sub_total = price * qty

    if not unit_id:
        print("[FAIL] first_stock_unit_id kosong")
        return failures + 1

    create_payload = {
        "first_stock": {
            "first_stock_date": SMOKE_DATE,
            "description": f"Smoke first stock {rnd()}"
        },
        "first_stock_items": [
            {
                "product_id": product_id,
                "unit_id": unit_id,
                "price": price,
                "qty": qty,
                "sub_total": sub_total,
                "expired_date": STOCK_TEST_EXPIRED_DATE
            }
        ]
    }

    status, headers, body = request("POST", "/api/first-stocks", token=branch_token, json_payload=create_payload)
    failures += 0 if expect_status("first_stock_create", status, 201) else 1
    payload = body_as_json(body) or {}
    data = payload.get("data") or {}
    fs_id = ((data.get("first_stock") or {}).get("id") if isinstance(data.get("first_stock"), dict) else None) or data.get("id")
    if not fs_id:
        print("[FAIL] first_stock_id kosong")
        return failures + 1

    status, product_after_create = get_product_detail(branch_token, product_id)
    failures += 0 if expect_status("stock_product_detail_after_create", status, 200) else 1
    if status == 200:
        stock_after_create = as_int(product_after_create.get("stock"), stock_before)
        failures += 0 if expect_equal("first_stock_stock_increment", stock_after_create, stock_before + qty) else 1

    status, headers, body = request("DELETE", f"/api/first-stocks/{fs_id}", token=branch_token)
    failures += 0 if expect_status("first_stock_delete", status, 200) else 1

    status, product_after_delete = get_product_detail(branch_token, product_id)
    failures += 0 if expect_status("stock_product_detail_after_delete", status, 200) else 1
    if status == 200:
        stock_after_delete = as_int(product_after_delete.get("stock"), stock_before)
        failures += 0 if expect_equal("first_stock_stock_rollback", stock_after_delete, stock_before) else 1

    return failures


def main():
    failures = 0
    cases = load_cases()

    status, headers, body = login_with_retry()
    if not expect_status("login", status, 200):
        return 1
    login_payload = body_as_json(body) or {}
    login_token = login_payload.get("data") or ""
    if not login_token:
        print("[FAIL] login_token kosong")
        return 1

    status, headers, body = request_with_retry("GET", "/api/menus", token=login_token)
    failures += 0 if expect_status("menus", status, 200) else 1

    status, headers, body = request_with_retry("GET", "/api/list_branches", token=login_token)
    failures += 0 if expect_status("list_branches", status, 200) else 1
    branches_payload = body_as_json(body) or {}
    branches = branches_payload.get("data") or []
    effective_branch = BRANCH_ID or (branches[0].get("branch_id") if branches else "")
    if not effective_branch:
        print("[FAIL] branch_id tidak ditemukan untuk set_branch")
        return 1

    status, headers, body = request("POST", "/api/set_branch", token=login_token, json_payload={"branch_id": effective_branch})
    failures += 0 if expect_status("set_branch", status, 200) else 1
    set_branch_payload = body_as_json(body) or {}
    branch_token = set_branch_payload.get("data") or ""
    if not branch_token:
        print("[FAIL] branch_token kosong")
        return 1

    failures += run_case_group(branch_token, cases.get("json_get_cases", []), "json_get_cases")
    failures += run_case_group(branch_token, cases.get("export_cases", []), "export_cases")
    failures += run_case_group(branch_token, cases.get("return_support_cases", []), "return_support_cases")

    if ENABLE_MUTATION:
        failures += run_lightweight_mutation_cases(branch_token)
    else:
        print("[info] mutation smoke dilewati karena SMOKE_ENABLE_MUTATION=0")

    if ENABLE_DEEP_STOCK:
        failures += run_deep_stock_case(branch_token)
    else:
        print("[info] deep stock smoke dilewati karena SMOKE_ENABLE_DEEP_STOCK=0")

    print("---")
    if failures:
        print(f"Smoke regression selesai dengan {failures} kegagalan")
        return 1

    print("Smoke regression selesai: semua check lulus")
    return 0


if __name__ == "__main__":
    sys.exit(main())
