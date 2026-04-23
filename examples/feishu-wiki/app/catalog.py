#!/usr/bin/env python3
# catalog.py - 下载知识库到本地

import json
import os
import time
import requests
from config import (
    BASE_DIR,
    SPACES_DIR,
    PAGE_SIZE,
    REQUEST_DELAY,
    space_dir,
    node_file,
    metadata_file,
    ensure_space_dirs,
)

os.makedirs(SPACES_DIR, exist_ok=True)
os.makedirs(BASE_DIR, exist_ok=True)

FEISHU_BASE = "https://open.feishu.cn"

# ─────────────────────────────────────────────
# HTTP 请求
# ─────────────────────────────────────────────

def feishu_get(path, params=None, _retry=True):
    """
    用 requests 直接请求飞书 API。
    token 由 token_store.get_token() 管理，过期自动刷新。
    遇到服务端报 token 过期时清缓存重试一次。
    """
    from token_store import get_token, clear_token
    token = get_token()
    url = FEISHU_BASE + path
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json; charset=utf-8",
    }
    try:
        resp = requests.get(url, headers=headers, params=params, timeout=30)
        resp.raise_for_status()
        data = resp.json()

        # 飞书 token 过期错误码
        if data.get("code") in (99991671, 99991672) and _retry:
            print("[TOKEN] 服务端报告过期，正在刷新...")
            clear_token()
            return feishu_get(path, params, _retry=False)

        return data
    except requests.HTTPError as e:
        print(f"[HTTP错误] {e} | {url} params={params}")
        return None
    except requests.RequestException as e:
        print(f"[请求失败] {e} | {url}")
        return None


# ─────────────────────────────────────────────
# 业务函数
# ─────────────────────────────────────────────

def get_spaces():
    spaces = []
    page_token = None
    while True:
        params = {"page_size": PAGE_SIZE}
        if page_token:
            params["page_token"] = page_token
        resp = feishu_get("/open-apis/wiki/v2/spaces", params)
        if not resp or resp.get("code") != 0:
            msg = resp.get("msg", "无响应") if resp else "无响应"
            print(f"[错误] 获取知识库列表失败: {msg}")
            break
        data = resp.get("data", {})
        spaces.extend(data.get("items", []))
        if not data.get("has_more"):
            break
        page_token = data.get("page_token")
    return spaces


def get_node_detail(node_token):
    if not node_token:
        return None
    return feishu_get("/open-apis/wiki/v2/spaces/get_node", {"token": node_token})


def save_json(data, filepath):
    os.makedirs(os.path.dirname(filepath), exist_ok=True)
    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    print(f"  [SAVE] {filepath}")


def clean_deleted_nodes(space_id, valid_tokens):
    nodes_dir = f"{space_dir(space_id)}/nodes"
    if not os.path.exists(nodes_dir):
        return
    valid_set = set(valid_tokens)
    deleted_count = 0
    for filename in os.listdir(nodes_dir):
        if filename.startswith("node_") and filename.endswith(".json"):
            token = filename[5:-5]
            if token not in valid_set:
                filepath = os.path.join(nodes_dir, filename)
                os.remove(filepath)
                deleted_count += 1
                print(f"  [DEL] {filepath}")
    if deleted_count > 0:
        print(f"  共清理 {deleted_count} 个已删除节点")


def fetch_nodes(space_id, parent_token, indent=0):
    page_token = None
    prefix = "  " * indent
    all_nodes = []
    node_index = []

    while True:
        params = {"page_size": PAGE_SIZE}
        if parent_token:
            params["parent_node_token"] = parent_token
        if page_token:
            params["page_token"] = page_token

        resp = feishu_get(f"/open-apis/wiki/v2/spaces/{space_id}/nodes", params)
        if not resp or resp.get("code") != 0:
            msg = resp.get("msg", "无响应") if resp else "无响应"
            print(f"{prefix}[错误] 获取节点列表失败: {msg}")
            break

        data = resp.get("data", {})
        items = data.get("items", [])

        for node in items:
            node_token = node.get("node_token")
            title = node.get("title", "未命名")
            print(f"{prefix}├─ {title}")

            all_nodes.append(node)
            node_index.append({
                "node_token": node_token,
                "title": title,
                "obj_type": node.get("obj_type"),
                "parent_token": node.get("parent_node_token"),
                "has_child": node.get("has_child"),
                "updated_at": node.get("updated_at"),
            })

            detail = get_node_detail(node_token)
            if detail and detail.get("code") == 0:
                node_data = {"node": node, "detail": detail.get("data", {})}
                save_json(node_data, node_file(space_id, node_token))
            elif detail:
                print(f"{prefix}  [WARN] 节点详情失败 ({node_token}): {detail.get('msg', '')}")

            if node.get("has_child"):
                child_nodes, child_index = fetch_nodes(space_id, node_token, indent + 1)
                all_nodes.extend(child_nodes)
                node_index.extend(child_index)

            time.sleep(REQUEST_DELAY)

        if not data.get("has_more"):
            break
        page_token = data.get("page_token")

    return all_nodes, node_index


# ─────────────────────────────────────────────
# 入口
# ─────────────────────────────────────────────

def main():
    print("=== 获取知识库列表 ===")
    spaces = get_spaces()
    if not spaces:
        print("未找到任何知识库")
        return

    target_space_ids = None
    try:
        from validate import load_contract
        contract_spaces = load_contract().get("spaces", [])
        if contract_spaces:
            target_space_ids = {s["id"] for s in contract_spaces}
            print(f"契约过滤：仅同步 {len(target_space_ids)} 个知识库")
    except (FileNotFoundError, ImportError):
        pass

    all_spaces = []

    for space in spaces:
        space_id = space.get("space_id")
        space_name = space.get("name")

        # 跳过未在契约中声明的空间（如果有契约的话）
        if target_space_ids and space_id not in target_space_ids:
            print(f"[SKIP] 未在契约中声明: {space_name} ({space_id})")
            continue

        print(f"\n{'=' * 50}")
        print(f"知识库: {space_name} ({space_id})")
        print(f"{'=' * 50}")

        ensure_space_dirs(space_id)

        nodes, node_index = fetch_nodes(space_id, "")
        print(f"\n共获取 {len(nodes)} 个节点")

        valid_tokens = [n["node_token"] for n in node_index]
        clean_deleted_nodes(space_id, valid_tokens)

        space_data = {
            "space": space,
            "nodes_count": len(nodes),
            "nodes": node_index,
            "synced_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
        }
        save_json(space_data, metadata_file(space_id))
        all_spaces.append({"space_id": space_id, "name": space_name, "nodes_count": len(nodes)})

    metadata = {
        "spaces": all_spaces,
        "total_spaces": len(all_spaces),
        "total_nodes": sum(s["nodes_count"] for s in all_spaces),
    }
    save_json(metadata, f"{BASE_DIR}/metadata.json")
    print(f"\n=== 完成，保存在 {BASE_DIR} ===")


if __name__ == "__main__":
    main()
