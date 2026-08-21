# openydt-cli (@openydt/openydt-cli)

艾科智泊智慧停车开放平台 CLI 的 npm 安装包 —— 为人和 AI Agent 而生(查费/缴费/车场/月票/电子券/设备等)。

```bash
npm i -g @openydt/openydt-cli      # 全局安装
openydt --version

# 或免安装试用
npx @openydt/openydt-cli --help
```

## 快速开始

```bash
openydt config set --profile demo --key <key> --secret <secret> --env test
openydt auth test
openydt schema getParkFee          # 查接口参数/枚举/示例
openydt trade get-park-fee --car-code 粤X12345 --park-code <车场>
```

## 全量分页导出

带 `pageNum` / `pageSize` 的只读查询可顺序获取全部记录并流式写成 NDJSON（默认每页间隔 500ms）：

```bash
openydt parking get-car-out-list \
  --park-code <车场> \
  --leave-start-time 20260601000000 \
  --leave-end-time 20260601235959 \
  --all-pages --out records.ndjson
```
