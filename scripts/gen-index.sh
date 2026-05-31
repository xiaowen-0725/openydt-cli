#!/usr/bin/env bash
# 由 catalog.json 生成 Agent 可读的接口索引(机器+人都能扫)。
set -euo pipefail
cd "$(dirname "$0")/.."
CAT=catalog/catalog.json
OUT=INTERFACE_INDEX.md
{
  echo "# openydt 接口索引"
  echo
  echo "> 由 \`make index\`(scripts/gen-index.sh)从 catalog.json 生成,**勿手改**。统计见 \`make counts\`。"
  echo "> 列:cmd | 方向(callable 可主动调 / webhook 平台推送) | 读写 | 是否一等命令 | 说明。"
  echo "> 一等命令用 \`openydt <域> <cmd>\` 调;callable 但非一等用 \`openydt api <cmd>\`;webhook 需自建接收端。"
  echo
  jq -r '
    .interfaces
    | group_by(.domain)[]
    | "## " + (.[0].domain) + " (" + (length|tostring) + ")\n\n"
      + "| cmd | 方向 | 读写 | 一等 | 说明 |\n|---|---|---|---|---|\n"
      + ( map("| `" + .cmd + "` | " + .direction + " | " + (.readwrite // "-")
              + " | " + (if .included then "✓" else "·" end)
              + " | " + ((.explain // "") | gsub("[\r\n|]"; " ") | .[0:50]) + " |") | join("\n") )
      + "\n"
  ' "$CAT"
} > "$OUT"
echo "wrote $OUT"
