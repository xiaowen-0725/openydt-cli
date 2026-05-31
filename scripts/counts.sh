#!/usr/bin/env bash
# 从 catalog.json 与 skills/ 生成权威计数,供 README/文档核对。
set -euo pipefail
cd "$(dirname "$0")/.."
CAT=catalog/catalog.json
echo "接口总数:     $(jq '.interfaces|length' $CAT)"
echo "一等命令:     $(jq '[.interfaces[]|select(.included==true and .direction=="callable")]|length' $CAT)"
echo "callable兜底: $(jq '[.interfaces[]|select(.included==false and .direction=="callable")]|length' $CAT)"
echo "webhook:      $(jq '[.interfaces[]|select(.direction=="webhook")]|length' $CAT)"
echo "技能数:       $(ls -d skills/openydt-* | wc -l | tr -d ' ')"
