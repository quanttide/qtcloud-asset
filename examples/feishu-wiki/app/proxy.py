#!/usr/bin/env python3
"""
proxy.py - 本地飞书 API 代理

用法：
  终端 A：python proxy.py
  终端 B：
    PowerShell: $env:LARK_ENDPOINT="http://localhost:7777"; python catalog.py
    CMD:        set LARK_ENDPOINT=http://localhost:7777 && python catalog.py
"""

import http.server
import urllib.request
import urllib.error
import json
import os
import sys

PROXY_PORT = int(os.getenv("PROXY_PORT", "7777"))
FEISHU_BASE = "https://open.feishu.cn"


class ProxyHandler(http.server.BaseHTTPRequestHandler):

    def log_message(self, fmt, *args):
        print(f"  [PROXY] {self.command} {self.path}")

    def _forward(self):
        from token_store import save_token

        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length) if content_length > 0 else None

        # 捕获 token
        auth = self.headers.get("Authorization", "")
        if auth.startswith("Bearer "):
            token = auth[7:].strip()
            if token:
                # 先用默认有效期存，后面从响应里拿到准确值再更新
                save_token(token, expires_in=7200)
                print(f"  [PROXY] ✅ 捕获 token: {token[:12]}...")

        # 转发到飞书，self.path 已含完整 query string
        target_url = FEISHU_BASE + self.path
        headers = {
            k: v for k, v in self.headers.items()
            if k.lower() not in ("host", "content-length")
        }
        headers["Host"] = "open.feishu.cn"

        req = urllib.request.Request(
            target_url, data=body, headers=headers, method=self.command
        )

        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                resp_body = resp.read()

                # 如果响应里有 expire 字段，用准确值更新 token 有效期
                try:
                    data = json.loads(resp_body)
                    expire = (
                        data.get("expire") or
                        data.get("data", {}).get("expire") or
                        data.get("expires_in")
                    )
                    if expire and auth.startswith("Bearer "):
                        save_token(auth[7:].strip(), expires_in=int(expire))

                    code = data.get("code", "?")
                    msg = data.get("msg", "")
                    items = len(data.get("data", {}).get("items", []))
                    print(f"  [PROXY] ← code={code} msg={msg!r} items={items}")
                except Exception:
                    pass

                self.send_response(resp.status)
                for k, v in resp.getheaders():
                    if k.lower() not in ("transfer-encoding",):
                        self.send_header(k, v)
                self.end_headers()
                self.wfile.write(resp_body)

        except urllib.error.HTTPError as e:
            body = e.read()
            self.send_response(e.code)
            self.end_headers()
            self.wfile.write(body)
            print(f"  [PROXY] ← HTTP {e.code}: {body[:200]}")
        except Exception as e:
            self.send_response(502)
            self.end_headers()
            self.wfile.write(str(e).encode())
            print(f"  [PROXY] ← 转发失败: {e}")

    do_GET = _forward
    do_POST = _forward
    do_PUT = _forward
    do_PATCH = _forward
    do_DELETE = _forward


if __name__ == "__main__":
    server = http.server.HTTPServer(("127.0.0.1", PROXY_PORT), ProxyHandler)
    print(f"🚀 飞书 API 代理已启动: http://127.0.0.1:{PROXY_PORT}")
    print(f"   转发目标: {FEISHU_BASE}")
    print(f"\n   PowerShell: $env:LARK_ENDPOINT='http://localhost:{PROXY_PORT}'; python catalog.py")
    print(f"   CMD:        set LARK_ENDPOINT=http://localhost:{PROXY_PORT} && python catalog.py")
    print(f"\n   Ctrl+C 停止\n")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n代理已停止")
