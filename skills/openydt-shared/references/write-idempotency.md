# 写操作幂等 / 重试安全

> ⚠️ 客户端对 404/连接重置/429/5xx **自动重试 + 退避**(见 shared「限速与重试」)。若写操作每次重试换新业务键,会**重复扣费 / 重复开通 / 重复入账**。本规则是硬约束。

## 硬规则(MUST / NEVER)
- **MUST 复用首次幂等键**:同一笔写操作的所有重试,**必须沿用首次生成的业务键**(`billCode` / `thirdBillCode` / `thirdpartyBillCode` / `uniqNo` / `transationNum`);键由调用方生成并保证全局唯一。
- **NEVER 重试换新键**:绝不为重试生成新键——平台靠该键去重,新键 = 平台视作新业务 = 重复扣费/开通。
- **`907`「账单已同步」= 幂等命中,按成功对账**:重发已成功的缴费/补缴收到 907,说明首次已生效,**不是失败、不要再发**,改为查询(`get-pay-bill` / `get-park-detail` / `get-online-vip-ticket`)确认。
- **重发前先查**:不确定首次是否生效时,先用对应读命令核对,确认未生效再用**同一键**重发。

## 各命令幂等键速查
| 域 | 写命令 | 幂等键 | 去重语义 |
| --- | --- | --- | --- |
| trade | `pay-park-fee` | `billCode` | 全局唯一,重试同键去重对账;907=已同步 |
| trade | `payback-batch` | 每条 `thirdBillCode` | 逐条去重 |
| trade | `set-points` / `set-prestore-for-c-park` | `thirdBillCode` | 同上 |
| ticket | `add-online-month-ticket` | `billCode` | 防重复开通 |
| ticket | `renew-online-vip-ticket` | `billCode` | 防重复续费扣费 |
| ticket | `deduct-month-ticket-config` | `thirdpartyBillCode` | 防重复扣减 |
| coupon | `sell-coupon` | `transationNum` | 防重复售券 |
| coupon | `create-fixed-coupon` | `uniqNo` | 防重复建券组 |
| parking | `update-wihhold-detail-bill` | `thirdBillCode` | 代扣订单去重 |
| parking | `supplement-parking-record-in` / `inventory-car` / `correct-*` | (无显式键) | 重试前先 `check-channel-exist-car`/`get-park-on-site-car`/`get-inventory-record` 确认,避免重复补录/重复盘点离场 |

> 无显式幂等键的写(补录/盘点/校正):重试前**先用读命令确认首次是否已生效**,别让客户端自动重试重复建记录。
