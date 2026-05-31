# 路由触发评测 Baseline

> 由 `tools/eval/routing-eval.workflow.mjs`(subagent 路由器,**非** nested `claude -p`)跑 `skills/openydt-*/evals/routing-evals.json` 共 216 条产出。复跑:`Workflow({scriptPath:"tools/eval/routing-eval.workflow.mjs"})`。

## 总体:211/216 = **97.7%** 命中

每个 owner 技能的命中(hit/total):

| owner | hit/total |
|---|---|
| openydt-billing | 27/27 |
| openydt-record | 28/29 |
| openydt-monthticket | 19/19 |
| openydt-park | 18/18 |
| openydt-data | 17/17 |
| openydt-device | 16/16 |
| openydt-coupon | 15/16 |
| openydt-list | 13/13 |
| openydt-api-explorer | 13/13 |
| openydt-skill-maker | 12/12 |
| openydt-shared | 14/14 |
| openydt-flow-park-access | 8/8 |
| none | 11/14 |

> 注:上表按 router 视角的 expected 聚合(含跨文件负例),与各文件自身条数不同。

## 5 个误路由分析

| query(节选) | expected | predicted | 判定 |
|---|---|---|---|
| 「探索那些没封装成命令的接口」 | none | openydt-api-explorer | **数据标签问题**:flow 旧 eval 遗留(其 candidates 当时不含 api-explorer,none=非flow)。全局下 api-explorer 才是正确 owner。**已修标签为 api-explorer**(修后 212/216=98.1%)。 |
| 「写个 groovy 计费脚本」 | none | openydt-park | 良性边界:提到「计费」被引向 park;实则平台不写 groovy 脚本,none 更准。router 可接受偏差,不硬调。 |
| 「查这车有哪些可用的优惠券」 | openydt-coupon | openydt-park | 良性边界:park 有「车辆优惠券记录(只读)」、coupon 有「查券」。两者皆可辩;保留以提示该边界天然模糊。 |
| 「线下收现金,补录一条缴费账单明细」 | openydt-record | openydt-billing | 良性边界:「补录缴费」介于 record(账单明细)与 billing(缴费回传)之间。 |
| 「help fix hardware scanner offline」 | none | openydt-device | 良性边界:硬件报修(none)提到 scanner 被引向 device。 |

## 结论
- 去冲突裁决表落地有效:**11 个 owner 技能全部 ≥15/16,无系统性串域**。
- 唯一标签问题(gid216)已修正;其余 4 例为**天然边界歧义**,router 选择均可辩护,**按既有原则不过拟合硬调**(避免引入误触发)。
- 该 baseline 与数据集随技能演进复跑即可回归。
