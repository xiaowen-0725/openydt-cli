# 设计:openydt skills 自动同步(对齐飞书 CLI)

> 日期:2026-05-30 · 状态:已批准架构,待落地
> 关联:`PROJECT_STATUS.md`(子 Agent 平台对齐)、飞书 CLI `internal/skillscheck` + `internal/selfupdate`

## 1. 背景与问题

openydt-cli 目前把技能(`skills/openydt-*/SKILL.md`,共 11 个)分发给 AI Agent 的唯一方式是 README 里的手动 `cp -r skills/openydt-* ~/.claude/skills/` —— **实际只覆盖 Claude Code 一个平台**,且需手动操作、不会随版本更新。

飞书 CLI 的做法是接入外部 **`skills` 包管理器**(`npx skills add larksuite/cli -g -y`),由它自动探测本机已装的 agent 并把技能分发过去(注册表支持 ~56 个平台:claude-code / codex / cursor / gemini-cli / opencode / openclaw / windsurf / qwen-code / …)。飞书的 `internal/skillscheck` 负责增量同步 + 版本漂移提示,在 `lark-cli update` 时触发实同步、在 `lark-cli` 启动时做本地版本比对并提示。

**已验证**:`npx skills add xiaowen-0725/openydt-cli --list` 能克隆仓库、识别全部 11 个技能,并自动探测当前 agent 非交互安装。技术路径成立。

### 核心需求(用户确认)

> **更新 openydt-cli 时,自动同步 skills。**

更新 openydt-cli 的方式 = `npm i -g @openydt/openydt-cli@latest`(或 `npm update -g`)。npm 在 **install 与 update 时都会重跑 `postinstall`**,因此把技能同步挂进 postinstall 即天然满足"update 时自动同步"。Go 二进制内的后台自动同步作为兜底,覆盖 postinstall 未跑到的边角场景。

## 2. 目标 / 非目标

### 目标
- `npm i/update -g @openydt/openydt-cli` 时,自动把 11 个技能同步到本机所有已装 agent(主路径,postinstall)。
- 二进制启动检测到"技能版本与二进制版本漂移"时,**后台静默自动重新同步**(兜底,覆盖 `--ignore-scripts`/非 npm 安装/上次失败)。
- 提供显式命令 `openydt skill sync [--force]` 供人/agent 手动同步。
- 全程 best-effort:技能同步失败 **绝不** 阻断 `npm install`,也 **绝不** 干扰任何 CLI 命令的 stdout(保护 agent 的 JSON 解析)。

### 非目标
- 不在 Go 里复刻飞书的增量 diff 规划(列官方 + 列本地 + 差集)。`npx skills add <repo> -g -y` 本身幂等更新全部,自己算增量是冗余。
- 不让二进制自更新(openydt 二进制由 npm postinstall 下载,保持现状)。
- 不新增 `openydt update` 命令(更新走 npm)。
- 不改技能内容、不改 `skills/` 布局(已与 `skills` 工具兼容)。

## 3. 架构(三条路径,一个 state 文件)

### ① 主路径 — npm postinstall(install + update 都触发)
`npm/install.js` 下载二进制后追加一步:
```
npx -y skills add xiaowen-0725/openydt-cli -g -y
```
- best-effort:失败/超时只打印 `[openydt] skills 同步失败(可稍后手动:openydt skill sync)`,**postinstall 退出码仍为 0**。
- 成功后写 `skills-state.json`:`{version: pkg.version, updated_at}`。

### ② 兜底 — Go 二进制后台自动同步(用户选定)
`cmd/root.go` 启动时经 `cobra.OnInitialize` 调 `skillsync.MaybeTrigger(cmdutil.Version)`:
1. `shouldSkip`:opt-out 环境变量 / CI / DEV / 非 release / 子进程标记 → 直接返回。
2. 读 state:`normalize(syncedVersion) == normalize(binaryVersion)`(去 `v`/`V` 前缀,同飞书)→ 已同步,返回。
3. 漂移或冷启动(无 state):
   - **防抖**:原子写 `last_attempt_version = binaryVersion` 进 state;若读到的 `last_attempt_version` 已等于当前版本 → 跳过(每版本最多 fork 一次,避免每次命令都跑、多进程并发风暴)。
   - `npx`/`node` 不可用 → 不 fork,存一条降级 notice(该版本只提示一次)。
   - 否则 **detached** 起子进程 `openydt skill sync --quiet`,stdout/stderr → `~/.config/openydt-cli/skills-sync.log`,父进程立即返回。

### ③ 手动命令 — `openydt skill sync [--force] [--quiet]`
- 前台跑 `npx -y skills add xiaowen-0725/openydt-cli -g -y` + 写 state,打印摘要。
- `--force`:全量重装(透传 `skills` 的全量行为)。
- `--quiet`:供 ② 的后台子进程用;子进程附带内部 env `OPENYDT_SKILL_SYNC_CHILD=1` 防止递归触发 ②。
- 失败时前台报错 + 退出码非 0(显式调用,该报错)。

## 4. 组件 / 包结构(对齐飞书,裁剪掉增量 diff)

```
internal/skillsync/
  state.go          SkillsState 读写(原子写, XDG 路径), 版本归一 normalizeVersion
  state_test.go
  skip.go           shouldSkip: OPENYDT_NO_SKILLS_SYNC / CI / DEV / 非 release / 子进程标记
  skip_test.go
  trigger.go        MaybeTrigger(version): 读 state → 比对 → 防抖 → detached spawn / 降级 notice
  trigger_test.go   注入假 spawner + 假 clock + 临时 state 目录;真 npx 不进单测
  spawn_unix.go     //go:build !windows   SysProcAttr{Setsid:true}
  spawn_windows.go  //go:build windows    CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS
  runner.go         Runner.SyncAll(): 跑 npx skills add ...;RunnerOverride 供测试注入
  runner_test.go
  notice.go         降级 notice 的存取(atomic.Pointer 或 state 字段), Message()
cmd/skill/
  skill.go          `openydt skill` 父命令 + `sync` 子命令(--force/--quiet)
  skill_test.go
```

变更点:
- `cmd/root.go`:`cobra.OnInitialize(func(){ skillsync.MaybeTrigger(cmdutil.Version) })` + `root.AddCommand(skillcmd.New(f))`。
- `npm/install.js`:下载后追加 best-effort `npx skills add` + 写 state。
- `internal/output/output.go`:JSON 模式注入 `_skills_notice`(**仅降级/失败场景**有值);table 模式 stderr 打一行。正常自动同步路径不产生任何用户可见噪声。

### state schema(`~/.config/openydt-cli/skills-state.json`,XDG-aware,原子写,0644)
```json
{
  "version": "0.1.1",
  "last_attempt_version": "0.1.1",
  "skills": ["openydt-billing", "openydt-park", "..."],
  "updated_at": "2026-05-30T12:00:00Z"
}
```
- `version`:上次**成功**同步对应的版本(漂移比对用此字段)。
- `last_attempt_version`:上次**尝试**(成功或失败)的版本,防抖用;失败时 `version` 不变、`last_attempt_version` 变 → 不会同版本反复 fork。
- `skills`:成功同步的技能名(信息性,手动 sync 时打印用;postinstall 写入时可省略)。
- 路径:`$XDG_CONFIG_HOME/openydt-cli/skills-state.json`,缺省 `~/.config/openydt-cli/`,复用 `internal/config` 的目录约定。

## 5. 数据流

**安装/更新(主路径)**
`npm i -g @openydt/openydt-cli@latest` → postinstall:下载二进制 → `npx skills add ... -g -y` → 写 state{version,updated_at}。

**日常命令(兜底自检)**
`openydt <任意命令>` → OnInitialize → MaybeTrigger:state.version == 二进制版本 → 无操作(零网络零子进程);漂移 → 写 last_attempt_version → detached fork `openydt skill sync --quiet`(输出进日志)→ 主命令照常执行返回。子进程跑完写 state.version。

**手动**
`openydt skill sync [--force]` → 前台 `npx skills add` → 写 state → 打印摘要。

## 6. 错误处理(best-effort,技能同步永不阻断主流程)

| 场景 | 行为 |
|---|---|
| postinstall `npx skills add` 失败/超时 | 打印提示;postinstall 退出码 0(不挂 `npm i`) |
| 兜底:`npx`/`node` 缺失 | 不 fork;存降级 notice(该版本提示一次) |
| 兜底:fork 失败 | 静默吞掉,主命令无感,下次启动重试 |
| 兜底:`npx skills add` 失败 | 写 `skills-sync.log`;不更新 `version`,`last_attempt_version` 已防抖,不风暴 |
| state 损坏/不可读 | 当冷启动处理,重新同步 |
| 手动 `skill sync` 失败 | 前台报错 + 退出码非 0 |
| 并发启动 | 原子写 `last_attempt_version` 作护栏,最多 fork 一次 |

## 7. 跨平台 detached spawn

子进程须脱离父会话,否则父退出会被带走。按平台拆文件,共用 `exec.Command(selfExe, "skill", "sync", "--quiet")`,**不调 `cmd.Wait()`**:
- `spawn_unix.go`(`!windows`):`SysProcAttr{Setsid: true}`,stdin=nil,stdout/stderr → log 文件。
- `spawn_windows.go`(`windows`):`SysProcAttr{CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS}`。
- `selfExe` 经 `os.Executable()` 解析。

## 8. 跳过规则 / opt-out

`shouldSkip(version)` 返回 true(对齐飞书 `skillscheck.shouldSkip`,换专属 env)。⚠️ openydt **没有** 飞书那样的 `internal/update.IsCIEnv`/`IsRelease` 辅助,因此 `skillsync` 包内自带两个小函数 `isCIEnv()`、`isReleaseVersion(v)`,不依赖外部包:
- `OPENYDT_NO_SKILLS_SYNC` 非空 → 跳过(用户彻底关闭自动同步)。
- `isCIEnv()`:`CI` / `GITHUB_ACTIONS` / `BUILD_NUMBER` 等任一非空 → 跳过。
- `isReleaseVersion(v)`:版本为 `dev`/`DEV`/空 → 跳过;或不匹配干净发布语义版本 `^v?\d+\.\d+\.\d+$` → 跳过(`git describe` 产出的 `v0.1.1-3-gabc-dirty` 这类带提交数/dirty 后缀的源码构建版本视为非 release,本地不打扰)。
- 子进程标记 `OPENYDT_SKILL_SYNC_CHILD=1`(防递归)。

## 9. 测试策略(对齐 CLAUDE.md「单测锚定」)

纯单元(`go test ./...` 全绿,`go vet ./...` 干净):
- `state`:读写往返、版本归一(`v0.1.1`/`V0.1.1`/`0.1.1` 等价)、损坏文件当冷启动、原子写。
- `skip`:各 env / CI / DEV / 非 release / 子进程标记矩阵。
- `trigger.MaybeTrigger`:注入假 spawner + 假 clock + 临时 state 目录,断言——已同步不 fork、漂移 fork 一次、防抖第二次不 fork、npx 缺失走降级 notice。
- `runner`:注入 override,断言命令行精确为 `-y skills add xiaowen-0725/openydt-cli -g -y`(及 `--list`/`-s` 变体若用)。
- `cmd/skill`:`sync --force` 透传、`--quiet` 抑制输出、失败退出码非 0。

**不进单测**:真实 `npx`(网络/环境依赖),与飞书一致用 override 隔离;真实分发由人工/E2E 另行验证。

## 10. 文档变更

- `README.md`:安装段把 `cp -r skills/openydt-* ~/.claude/skills/` 改为 `npx skills add xiaowen-0725/openydt-cli -g -y`(并说明 `npm i -g` 已自动同步、无需手动);保留 npm 为主安装方式。
- `CLAUDE.md`:新增「技能同步」小节(主路径 postinstall / 兜底二进制 / 手动命令 / opt-out env / state 文件位置)。
- `PROJECT_STATUS.md`:在"待办"里把"子 Agent 平台对齐"标记为已落地(实现合并后更新)。

## 11. 风险 / 已知点

- **版本号一致性**:postinstall 的 `pkg.version`(如 `0.1.1`)需与二进制 `cmdutil.Version`(git tag,可能 `v0.1.1`)对得上;`normalizeVersion` 去前缀后比对,已覆盖。发布流程须保证 npm 包版本 == GitHub tag == 二进制 ldflags 版本(现有 release 流水线已如此)。
- ~~**首发 npm 阻塞**~~(已解除,2026-05-30):`@openydt/openydt-cli@0.1.1` 已发布到 npm,端到端"`npm i -g @openydt/openydt-cli` 自动同步"现可直接真机验证。⚠️ 注意:postinstall 加入 `npx skills add` 后,需发一个**新 npm 版本**(如 `0.1.2`)才能让该步随安装/更新对外生效;`0.1.1` 的已装用户由二进制后台兜底路径覆盖。
- **`skills` 工具行为变化**:`npx skills` 为外部依赖,其 CLI 参数若变更需跟进;用 `-y` 固定非交互,Runner 集中一处便于改。
- **Windows detached**:`DETACHED_PROCESS` 行为需在 Windows 实测;若不可靠,降级为"只提示不后台跑"(Windows 下退回 notice 模式)。
