#!/usr/bin/env python3
"""
token_store.py - token 的读写、过期管理和自动刷新

.lark_token 格式：
  {"token": "u-xxx", "refresh_token": "ur-xxx", "expires_at": 1713600000.0}

App ID + Secret 放在 .env：
  FEISHU_APP_ID=cli_xxxxxxxx
  FEISHU_APP_SECRET=xxxxxxxxxxxxxxxx
"""

import json
import os
import time
import requests

TOKEN_FILE = os.path.join(os.path.dirname(__file__), ".lark_token")
EXPIRE_BUFFER = 300  # 提前 5 分钟视为过期


def save_token(token: str, expires_in: int = 7200, refresh_token: str = None):
    """保存 token，用原子写入避免读到损坏数据。"""
    data = {
        "token": token,
        "refresh_token": refresh_token,
        "expires_at": time.time() + expires_in - EXPIRE_BUFFER,
    }
    tmp = TOKEN_FILE + ".tmp"
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(data, f)
    os.replace(tmp, TOKEN_FILE)


def load_token() -> str | None:
    """读取 token，过期返回 None。"""
    if not os.path.exists(TOKEN_FILE):
        return None
    try:
        with open(TOKEN_FILE, encoding="utf-8") as f:
            data = json.load(f)
        if time.time() < data.get("expires_at", 0):
            return data["token"]
        print("[TOKEN] 已过期")
        return None
    except (json.JSONDecodeError, KeyError):
        os.remove(TOKEN_FILE)
        return None


def refresh_token() -> str:
    """
    用 refresh_token 换新的 user_access_token。
    如果 refresh_token 也过期，则需要重新授权获取新 code。
    """
    from dotenv import load_dotenv
    load_dotenv()

    if not os.path.exists(TOKEN_FILE):
        raise RuntimeError(
            "未找到 token 文件，需要重新授权。\n"
            "请在浏览器打开：\n"
            "https://open.feishu.cn/open-apis/authen/v1/authorize"
            "?app_id=cli_a96f60492c3b1cbb"
            "&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback"
            "&state=xyz"
        )

    with open(TOKEN_FILE, encoding="utf-8") as f:
        data = json.load(f)

    refresh_token_str = data.get("refresh_token")
    if not refresh_token_str:
        raise RuntimeError("没有 refresh_token，需要重新授权")

    app_id = os.getenv("FEISHU_APP_ID")
    app_secret = os.getenv("FEISHU_APP_SECRET")

    # 用 refresh_token 换 user_access_token
    resp = requests.post(
        "https://open.feishu.cn/open-apis/authen/v1/oidc/refresh_access_token",
        json={
            "grant_type": "refresh_token",
            "refresh_token": refresh_token_str,
            "app_id": app_id,
            "app_secret": app_secret,
        },
        timeout=15,
    )
    resp.raise_for_status()
    result = resp.json()

    if result.get("code") != 0:
        raise RuntimeError(f"刷新 token 失败: {result}")

    token = result["data"]["access_token"]
    new_refresh = result["data"].get("refresh_token")
    expires_in = result["data"].get("expires_in", 7200)

    save_token(token, expires_in=expires_in, refresh_token=new_refresh)
    print(f"[TOKEN] 已刷新 user token，有效期 {expires_in // 60} 分钟")
    return token


def get_token() -> str:
    """
    对外唯一入口：有效 token 直接返回，过期自动刷新。
    catalog.py 只需要调这一个函数。
    """
    return load_token() or refresh_token()


def clear_token():
    if os.path.exists(TOKEN_FILE):
        os.remove(TOKEN_FILE)
        print("[TOKEN] 已清除")