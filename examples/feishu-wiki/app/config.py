# config.py - 飞书知识库下载配置

import os
from dotenv import load_dotenv

load_dotenv()

BASE_DIR = os.getenv("BASE_DIR", "./data/feishu_wiki")
SPACES_DIR = f"{BASE_DIR}/spaces"
PAGE_SIZE = int(os.getenv("PAGE_SIZE", "50"))
REQUEST_DELAY = float(os.getenv("REQUEST_DELAY", "0.2"))

def space_dir(space_id):
    return f"{SPACES_DIR}/space_{space_id}"

def node_file(space_id, node_token):
    return f"{space_dir(space_id)}/nodes/node_{node_token}.json"

def metadata_file(space_id):
    return f"{space_dir(space_id)}/metadata.json"

def ensure_space_dirs(space_id):
    os.makedirs(space_dir(space_id), exist_ok=True)
    os.makedirs(os.path.dirname(node_file(space_id, "")), exist_ok=True)
