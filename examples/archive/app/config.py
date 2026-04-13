#!/usr/bin/env python3
"""示例归档配置"""

from pathlib import Path

ROOT = Path(__file__).parents[3]
FIXTURES = ROOT / "assets" / "fixtures"

JOURNAL = FIXTURES / "docs" / "journal"
ARCHIVE = FIXTURES / "docs" / "archive"
