---
name: openydt-device
version: 1.0.0
description: "停车场设备控制：远程开关闸/开闸/关闸、修改闸机模式(常开/正常)、显示屏播报/语音播报、下发默认屏显内容、手动抓拍、云端扫码机扫码/停止扫码/语音/更新配置、获取云端设备状态。触发词:开闸/关闸/抬杆/落杆/远程开门/道闸控制/闸机常开/一体机模式/屏显/显示屏/LED播报/语音播报/喊话/抓拍/拍照/扫码机/二维码扫码/设备状态/在线状态/通道控制/智慧岗亭"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt device --help"
---

> **CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**：其中包含认证 / profile / 签名 / 状态码 / 限速 / 安全等通用约定。未读基座规范不要执行任何命令。

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

## 示例

```bash
# 1) 开闸(写,高危):先预览
openydt device op-gate \
  --park-code 2KNTYVWC --channel-code 001 --op-type 0 \
  --operator operator --operate-time "2017-09-11 14:04:04" --dry-run
# 预览无误后真正执行
openydt device op-gate \
  --park-code 2KNTYVWC --channel-code 001 --op-type 0 \
  --operator operator --operate-time "2017-09-11 14:04:04" --yes

# 2) 显示屏 + 语音播报(写,需 --yes;show/voice 至少一个)
openydt device op-show-voice \
  --park-code 2KNTYVWC --channel-code 001 \
  --show show --voice voice \
  --operator operator --operate-time "2017-09-11 14:04:04" --yes

# 3) 查询云端扫码机状态(读,无需 --yes)
openydt device get-cloud-equip-status --equip-type 3 --client-id 3571F003
```
