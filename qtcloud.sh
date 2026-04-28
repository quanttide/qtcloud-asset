#!/bin/bash
# QtCloud CLI 便捷脚本
cd "$(dirname "$0")" && PYTHONPATH=src python -m src.cli.main "$@"
