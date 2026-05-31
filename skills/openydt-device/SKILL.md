---
name: openydt-device
version: 1.0.2
description: "设备控制域(device)：远程开关闸、改闸机模式(常开/正常)、向显示屏下发文字与语音播报、下发默认屏显、手动抓拍、云端扫码机扫码/停止/语音/更新配置、查云端设备在线状态。当用户要现场对道闸/岗亭/屏显/扫码机下指令或查设备状态时使用——高危现场运维，写操作需 --yes、建议先 --dry-run。注意：本域负责向设备「下发」屏显/播报；查某车「应显示什么」内容在 park 域。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt device --help"
---

# openydt-device — 设备控制域 (device)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

当用户意图是「对停车场现场设备下指令或读设备状态」时使用本技能,例如:远程开闸/关闸、把闸机切到常开模式、给显示屏推文字、让设备语音播报/喊话、抓拍车辆图片、控制云端扫码机(开始扫码/停止扫码/语音/更新配置)、查询设备在线状态。

意图路由:
- 「开闸 / 关闸 / 抬杆 / 落杆 / 远程开门」→ 传统/云场用 `op-gate`;纯云场按通道 ID 开关用 `cloud-open-gate`。
- 「常开 / 恢复正常 / 改闸机模式」→ `change-channel-mode`。
- 「显示屏 / LED 显示 / 语音播报 / 喊话」→ `op-show-voice`;设置默认屏显图片轮播 → `set-default-screen`。
- 「抓拍 / 拍照」→ `channel-snap`。
- 「扫码机 / 二维码支付设备」→ 扫码 `cloud-scan-qr-code`、停止 `cloud-stop-scan-code`、语音 `cloud-scan-voice`、更新配置 `cloud-scan-update-config`。
- 「设备状态 / 是否在线」→ `get-cloud-equip-status`。

> 不涉及现场设备的查询(车辆在场、订单、月票、券)请改用对应业务域技能(`openydt parking` / `openydt trade` / `openydt ticket` / `openydt coupon` 等)。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 修改一体机闸机模式 | `openydt device change-channel-mode` | 写 | parkCode, channelCode, mode(0常开/1正常) |
| 手动抓拍 | `openydt device channel-snap` | 写 | parkCode, channelCode, operator(可选) |
| 云场通道开/关闸 | `openydt device cloud-open-gate` | 写 | parkCode, channelId, opType(0开/1关) |
| 扫码机下发扫码 | `openydt device cloud-scan-qr-code` | 写 | parkCode, scanMachineId, deviceType, timeLength, voiceType |
| 扫码机更新配置 | `openydt device cloud-scan-update-config` | 写 | scanMachineId, channelId |
| 扫码机语音播报 | `openydt device cloud-scan-voice` | 写 | parkCode, scanMachineId, voiceType, voiceNum, voiceInterval |
| 扫码机停止扫码 | `openydt device cloud-stop-scan-code` | 写 | parkCode, scanMachineId, deviceType(可选) |
| 获取云端设备状态 | `openydt device get-cloud-equip-status` | 读 | equipType(0一体机/2卡机/3扫码机), clientId |
| 开/关闸 | `openydt device op-gate` | 写 | parkCode, channelCode, opType(0开/1关), operator, operateTime, carNo(可选) |
| 显示屏/语音播报 | `openydt device op-show-voice` | 写 | parkCode, channelCode, show/voice(至少一个), qrCode(可选), operator, operateTime |
| 设置默认屏显内容 | `openydt device set-default-screen` | 写 | parkCode, deviceType, channelCode, templateData.imageArray |

> 除 `get-cloud-equip-status`(读)外,以上全部为**写操作**,执行时必须加 `--yes` 确认。

> ⚠️ **`channel-snap` 需要通道上有抓拍设备**:若返回 `resultCode=908 找不到设备`,说明该通道没有抓拍设备,请换其他通道再试。

> 本表未列、但属设备域的可调用接口(`setLeavePrompt`/`removeLeavePrompt`/`setShowMsg`/`setVipShowMsg`/`addMidAccount`/`scanMachineFlow`),用 `openydt api <cmd> --body '{...}'` 调用(写 `--yes`、先 `--dry-run`);详见 [[openydt-api-explorer]]。

## 业务流程

设备控制属于直接作用于现场硬件的高危操作。标准顺序:**先定位设备 → 干预前可先查状态 → 用 `--dry-run` 预览 → 确认无误再加 `--yes` 真正下发**。务必把前序命令响应里的字段透传到后续命令入参,不要手填臆测值。

1. **定位通道 / 设备**:在停车场域(`openydt parking` / `openydt park`)拿到 `parkCode`、通道 `channelCode` 或 `channelId`、设备 `scanMachineId` / `clientId`。这些是后续所有设备命令的入参来源。

2. **(可选)先查设备在线状态**:对云端设备先跑 `get-cloud-equip-status`(`equipType` + `clientId`),响应里设备在线/状态字段确认设备可用,再决定是否下发指令。`clientId` 即扫码机的 `scanMachineId`。

3. **预览写指令(高危,务必先 dry-run)**:对 `op-gate`、`cloud-open-gate`、`op-show-voice`、`change-channel-mode`、`set-default-screen`、`channel-snap` 及扫码机系列命令,先用 `--dry-run` 查看将要发送的请求体与目标,核对 `parkCode` / `channelCode` / `channelId` / `opType` 等无误。

4. **确认下发**:复核 dry-run 输出后,把同一组参数加 `--yes` 正式执行。
   - 开关闸:传统场/云场用 `op-gate`(`channelCode` + `opType`);纯云场按通道 ID 用 `cloud-open-gate`(用第 1 步拿到的 `channelId` 而非 channelCode)。
   - 闸机常开:`change-channel-mode`(`mode=0` 常开 / `mode=1` 正常)。
   - 屏显/喊话:`op-show-voice`(`show` 与 `voice` 至少给一个);扫码机喊话用 `cloud-scan-voice`(`scanMachineId` 取自第 1/2 步)。
   - 扫码流程:`cloud-scan-qr-code` 开始扫码 →(需要时)`cloud-scan-voice` 提示 → `cloud-stop-scan-code` 停止,三者复用同一 `scanMachineId`。

5. **核对结果**:看返回状态码(参见基座规范)。开关闸等动作可再次 `get-cloud-equip-status` 或在停车场域复查在场记录确认生效。

> 📖 **结果解读**:开关闸/抓拍 `status=1` 表示**指令已下发**,不等于物理动作完成——以 `get-cloud-equip-status` 在线状态或停车场域复查为准。三层判读见 [[openydt-shared]] 的 [`references/result-reading-sop.md`](../openydt-shared/references/result-reading-sop.md)。
> 🔑 **重复下发风险**:开关闸/抓拍/扫码**无幂等键**,网关 404 触发的自动重试可能重复开闸/抓拍。高危写**先 dry-run、单次执行**,不确定是否生效用 `get-cloud-equip-status` 复查而非盲目重发。详见 [[openydt-shared]] 的 [`references/write-idempotency.md`](../openydt-shared/references/write-idempotency.md)。

## 错误自愈速查（设备）

| 现象 | 含义 | 恢复动作 |
| --- | --- | --- |
| `resultCode=908 找不到设备` | 该通道无对应设备(抓拍/扫码机) | 换有设备的通道;扫码机核对 scanMachineId |
| `status=7` 参数不完整 | channelId 与 channelCode 用错(云场用 channelId) | 纯云场用 channelId、传统/云用 channelCode;按 schema 核对 |
| 下发后设备无反应 | 设备离线 | `get-cloud-equip-status` 查在线再下发 |

> 通用码/退出码/重试见 [[openydt-shared]];结果解读(三层判读/金额/空结果)见其 `references/result-reading-sop.md`;写幂等规则见其 `references/write-idempotency.md`。

## 示例

> 示例 parkCode/时间为文档化测试值(仅 test);照抄历史 sampleBody 会撞无效车场/过期时间。`--client-id` 的 `3571F003` 取自设备查询,实际换真实设备 ID。

```bash
# 1) 开闸(写,高危):先预览
openydt device op-gate \
  --park-code 1ZS7H5PQH9 --channel-code 001 --op-type 0 \
  --operator operator --operate-time "2026-06-01 14:04:04" --dry-run
# 预览无误后真正执行
openydt device op-gate \
  --park-code 1ZS7H5PQH9 --channel-code 001 --op-type 0 \
  --operator operator --operate-time "2026-06-01 14:04:04" --yes

# 2) 显示屏 + 语音播报(写,需 --yes;show/voice 至少一个)
openydt device op-show-voice \
  --park-code 1ZS7H5PQH9 --channel-code 001 \
  --show "欢迎光临" --voice "请缴费" \
  --operator operator --operate-time "2026-06-01 14:04:04" --yes

# 3) 查询云端扫码机状态(读,无需 --yes)
openydt device get-cloud-equip-status --equip-type 3 --client-id 3571F003
```
