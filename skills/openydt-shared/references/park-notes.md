# 车场经验（park-notes）

按车场积累的经验存在 **openydt 配置目录**下的 `park-notes/{parkCode}.{env}.md`（默认 `~/.config/openydt-cli/park-notes/`，或 `$XDG_CONFIG_HOME/openydt-cli/park-notes/`；父目录即 `openydt config path` 所示文件的所在目录）。例如 `PTD2YBBZ.test.md`。**一车场一环境一文件，物理隔离 test/dev/prod，避免把 test 经验误用到 prod、prod 文件天然不含 test 的 PII 车牌。** **不要**放进技能目录——技能会被 `npx skills` 同步覆盖，经验会丢。

**任务开始前（回忆）**：
1. 先确认当前环境：`openydt config list` 看当前 profile 的 env，或按 `--env` 显式指定。
2. `ls ~/.config/openydt-cli/park-notes/`（目录不存在或为空属正常）。文件名形如 `{parkCode}.{env}.md`；必要时读各文件 frontmatter 的 `aliases` 做车场名/俗称匹配。
3. 用户指明目标车场（parkCode 或别名）后，读取匹配 **该 parkCode 且 env 与当前环境一致** 的文件（若有），据此选择签名版本、必填字段、避开已知陷阱、复用常用车牌。
4. 经验标注 `updated` 日期，**当"可能有效的提示"而非保证**；按经验操作若失败 → 回退通用流程，并**更新**该文件对应条目。

**openydt 命令成功（status=1）后（沉淀）**：
5. 若发现该车场值得记录的**已验证**新事实（可用签名版本、必填 body 字段、某接口在该环境 nodata、计费模式、稳定的关联 ID、常用车牌），主动**追加/更新**到 `park-notes/{parkCode}.{env}.md`（当前环境）对应小节并刷新 frontmatter `updated`。文件不存在则按下方模板创建（先 `mkdir -p` 该目录）。只记录在**当前环境**验证过的事实。
6. **只写经过验证的事实，不写猜测。**
7. **隐私红线**：车牌是 PII —— `常用车牌` 仅在 **test/dev** 的文件记录；**`{parkCode}.prod.md` 不写真实车牌**（可写"该车场有 VIP 车类型"这类非 PII 事实）。frontmatter 的 `env` 字段须与文件名中的 env 一致。

文件格式（文件名 `{parkCode}.{env}.md`，例如 `PTD2YBBZ.test.md`）：
```markdown
---
parkCode: PTD2YBBZ
aliases: [智汇云测试车场]
traderCode: O563CNQTNZVY
env: test
updated: 2026-05-30
---
## 车场特征
- 云车场；test 环境授权；按时段计费
## 有效调用
- getParkFee：v2 签名调通，body 仅需 carCode
## 已知陷阱
- dataAnalysis/* 在 test 多为 nodata
## 常用车牌
- 桂566666（普通）、粤HH7772（VIP）   # 仅 test/dev
```
