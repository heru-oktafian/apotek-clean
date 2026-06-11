#!/usr/bin/env python3
import json
import os
import random
import string
import sys
import time
import urllib.parse
import urllib.request
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
STARTUP_RETRIES = int(os.getenv("SMOKE_STARTUP_RETRIES", "12"))
STARTUP_DELAY_SECONDS = float(os.getenv("SMOKE_STARTUP_DELAY_SECONDS", "1.5"))


def rnd(n=5):
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=n))


def request(method, path, token=None, json_body=None):
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    data = None
    if json_body is not None:
        data = json.dumps(json_body).encode("utf-8")
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


def json_body(body):
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


def login_with_retry():
    last = (0, {}, b"")
    for attempt in range(1, STARTUP_RETRIES + 1):
        status, headers, body = request("POST", "/api/login", json_body={"username": USERNAME, "password": PASSWORD})
        last = (status, headers, body)
        if status == 200:
            if attempt > 1:
                print(f"[info] server siap pada percobaan ke-{attempt}")
            return last
        if status != 0:
            return last
        time.sleep(STARTUP_DELAY_SECONDS)
    return last


def run_json_get_cases(branch_token):
    failures = 0
    json_cases = [
        ("profile", "/api/profile", 200),
        ("expenses_list", f"/api/expenses?page=1&search=&month={MONTH}", 200),
        ("another_incomes_list", f"/api/another-incomes?page=1&search=&month={MONTH}", 200),
        ("first_stocks_list", f"/api/first-stocks?page=1&search=&month={MONTH}", 200),
        ("buy_returns_list", f"/api/buy-returns?page=1&search=&month={MONTH}", 200),
        ("sale_returns_list", f"/api/sale-returns?page=1&search=&month={MONTH}", 200),
        ("cmb_purchases", f"/api/cmb-purchases?search=&month={MONTH}", 200),
        ("cmb_sales", f"/api/cmb-sales?search=&month={MONTH}", 200),
        ("dashboard_daily_profit", "/api/dashboard/daily-profit-report", 200),
        ("neraca_saldo", f"/api/report/neraca-saldo?month={MONTH}", 200),
    ]

    for name, path, expected in json_cases:
        status, headers, body = request("GET", path, token=branch_token)
        failures += 0 if expect_status(name, status, expected) else 1
        if status == expected:
            failures += 0 if expect_content_type(name, headers, "application/json") else 1

    return failures


def run_export_cases(branch_token):
    failures = 0
    export_cases = [
        ("expenses_pdf", f"/api/expenses/pdf?month={MONTH}", "application/pdf"),
        ("expenses_excel", f"/api/expenses/excel?month={MONTH}", "spreadsheetml"),
        ("another_incomes_pdf", f"/api/another-incomes/pdf?month={MONTH}", "application/pdf"),
        ("another_incomes_excel", f"/api/another-incomes/excel?month={MONTH}", "spreadsheetml"),
        ("first_stocks_pdf", f"/api/first-stocks/pdf?month={MONTH}", "application/pdf"),
        ("first_stocks_excel", f"/api/first-stocks/excel?month={MONTH}", "spreadsheetml"),
        ("daily_assets_excel", f"/api/daily-assets/excel?month={MONTH}", "spreadsheetml"),
    ]

    for name, path, expected_content_type in export_cases:
        status, headers, body = request("GET", path, token=branch_token)
        failures += 0 if expect_status(name, status, 200) else 1
        if status == 200:
            failures += 0 if expect_content_type(name, headers, expected_content_type) else 1

    return failures


def run_return_support_cases(branch_token):
    failures = 0
    support_cases = [
        ("cmb_prod_buy_returns", f"/api/cmb-prod-buy-returns?purchase_id={urllib.parse.quote(PURCHASE_ID)}", 200),
        ("cmb_prod_sale_returns", f"/api/cmb-prod-sale-returns?sale_id={urllib.parse.quote(SALE_ID)}", 200),
        ("buy_returns_empty_month", f"/api/buy-returns?page=1&search=&month={EMPTY_MONTH}", 200),
        ("sale_returns_empty_month", f"/api/sale-returns?page=1&search=&month={EMPTY_MONTH}", 200),
        ("cmb_purchases_empty_month", f"/api/cmb-purchases?search=&month={EMPTY_MONTH}", 200),
        ("cmb_sales_empty_month", f"/api/cmb-sales?search=&month={EMPTY_MONTH}", 200),
        ("cmb_prod_buy_returns_empty", "/api/cmb-prod-buy-returns?purchase_id=PUR000000000000", 200),
        ("cmb_prod_sale_returns_empty", "/api/cmb-prod-sale-returns?sale_id=SAL000000000000", 200),
    ]

    for name, path, expected in support_cases:
        status, headers, body = request("GET", path, token=branch_token)
        failures += 0 if expect_status(name, status, expected) else 1
        if status == expected:
            failures += 0 if expect_content_type(name, headers, "application/json") else 1
            failures += 0 if expect_data_is_list(name, json_body(body) or {}) else 1

    return failures


def run_lightweight_mutation_cases(branch_token):
    failures = 0

    expense_payload = {
        "expense_date": SMOKE_DATE,
        "description": f"Smoke expense {rnd()}",
        "total_expense": 12345,
        "payment": "paid_by_cash",
    }
    status, headers, body = request("POST", "/api/expenses", token=branch_token, json_body=expense_payload)
    failures += 0 if expect_status("expense_create", status, 200) else 1
    expense_data = json_body(body) or {}
    expense_id = ((expense_data.get("data") or {}).get("id") if isinstance(expense_data, dict) else None)
    if not expense_id:
        print("[FAIL] expense_id kosong")
        return failures + 1

    expense_update_payload = dict(expense_payload)
    expense_update_payload["description"] = expense_payload["description"] + " updated"
    expense_update_payload["total_expense"] = 23456
    status, headers, body = request("PUT", f"/api/expenses/{expense_id}", token=branch_token, json_body=expense_update_payload)
    failures += 0 if expect_status("expense_update", status, 200) else 1

    status, headers, body = request("DELETE", f"/api/expenses/{expense_id}", token=branch_token)
    failures += 0 if expect_status("expense_delete", status, 200) else 1

    income_payload = {
        "income_date": SMOKE_DATE,
        "description": f"Smoke income {rnd()}",
        "total_income": 54321,
        "payment": "paid_by_cash",
    }
    status, headers, body = request("POST", "/api/another-incomes", token=branch_token, json_body=income_payload)
    failures += 0 if expect_status("income_create", status, 200) else 1
    income_data = json_body(body) or {}
    income_id = ((income_data.get("data") or {}).get("id") if isinstance(income_data, dict) else None)
    if not income_id:
        print("[FAIL] income_id kosong")
        return failures + 1

    income_update_payload = dict(income_payload)
    income_update_payload["description"] = income_payload["description"] + " updated"
    income_update_payload["total_income"] = 65432
    status, headers, body = request("PUT", f"/api/another-incomes/{income_id}", token=branch_token, json_body=income_update_payload)
    failures += 0 if expect_status("income_update", status, 200) else 1

    status, headers, body = request("DELETE", f"/api/another-incomes/{income_id}", token=branch_token)
    failures += 0 if expect_status("income_delete", status, 200) else 1

    return failures


def main():
    failures = 0

    status, headers, body = login_with_retry()
    if not expect_status("login", status, 200):
        return 1
    login_payload = json_body(body) or {}
    login_token = login_payload.get("data") or ""
    if not login_token:
        print("[FAIL] login_token kosong")
        return 1

    status, headers, body = request("GET", "/api/menus", token=login_token)
    failures += 0 if expect_status("menus", status, 200) else 1

    status, headers, body = request("GET", "/api/list_branches", token=login_token)
    failures += 0 if expect_status("list_branches", status, 200) else 1
    branches_payload = json_body(body) or {}
    branches = branches_payload.get("data") or []
    effective_branch = BRANCH_ID or (branches[0].get("branch_id") if branches else "")
    if not effective_branch:
        print("[FAIL] branch_id tidak ditemukan untuk set_branch")
        return 1

    status, headers, body = request("POST", "/api/set_branch", token=login_token, json_body={"branch_id": effective_branch})
    failures += 0 if expect_status("set_branch", status, 200) else 1
    set_branch_payload = json_body(body) or {}
    branch_token = set_branch_payload.get("data") or ""
    if not branch_token:
        print("[FAIL] branch_token kosong")
        return 1

    failures += run_json_get_cases(branch_token)
    failures += run_export_cases(branch_token)
    failures += run_return_support_cases(branch_token)

    if ENABLE_MUTATION:
        failures += run_lightweight_mutation_cases(branch_token)
    else:
        print("[info] mutation smoke dilewati karena SMOKE_ENABLE_MUTATION=0")

    print("---")
    if failures:
        print(f"Smoke regression selesai dengan {failures} kegagalan")
        return 1

    print("Smoke regression selesai: semua check lulus")
    return 0


if __name__ == "__main__":
    sys.exit(main())
