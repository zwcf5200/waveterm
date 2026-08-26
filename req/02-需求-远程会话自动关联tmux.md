# 需求：远程会话自动关联 tmux（Wave Terminal）

> 需求讨论记录 v3（已定方案 A）
> 目标：在 waveterm 中打开远程 SSH block 时，自动 `tmux new -A -t <会话名>` 进入上次的 tmux 会话，免去手动输入。

---

## 0. 背景与决策前提（必须先读）

- **已放弃 waveterm 自带 durable session**：官方 durable 功能有 bug，运行时间一长会卡死。
- **已改用 tmux 保活**：远程进程由 tmux 守护，SSH/Wave 断开不影响会话。
- **当前痛点**：每次在 Wave 打开远程 block，需手动输入 `tmux new -A -t omlx-11335` 之类命令重新 attach，繁琐。
- **期望**：打开远程 block 时**自动进入上次的 tmux 会话**，无需手动敲命令。

---

## 1. 需求定义

### 1.1 核心需求

> 在 waveterm 中打开一个"远程会话"类型的 terminal block 时，shell 启动后**自动执行 `tmux new -A -t <会话名>`**，attach 到对应 tmux 会话（不存在则创建，存在则恢复）。

### 1.2 明确的设计取向（已和用户确认）

1. **用「会话名」由独立 meta key 携带**，默认可自由命名（如 `omlx-11335`）。
2. **支持重命名**：改该 meta 值即换会话名。
3. **标识"远程会话"**：block 具备该 meta key 即视为"tmux 保活的远程会话"。
4. **自动关联**：重启/恢复 block 时自动 attach。

### 1.3 已确认的技术事实（影响实现）

| 事实 | 依据 |
|------|------|
| block 启动时支持执行自定义 init script | `shellcontroller.go:817 getCustomInitScript`，meta key `cmd:initscript`(`.zsh`/`.sh`/`.bash`) |
| swap token 是 shell 启动的"环境+脚本"载体 | `blockcontroller.go:459 makeSwapToken`，`token.Env` 已注 `WAVETERM_BLOCKID`/`WAVETERM_CONN` 等 |
| init script 内容注入位置 | `makeSwapToken` 里 `token.ScriptText = getCustomInitScript(...)` |
| ⚠️ **block 本身没有「Name」字段** | `wtype.go:294 Block` 只有 OID/Meta/ColorKey，**Workspace 才有 Name，Tab 才有 Name** |
| 前端可读写 block meta | `getBlockMetaKeyAtom` / waveEnv 机制 |
| 进入 tmux 后 Wave shell integration 被禁用 | `zsh_zshrc.sh:30 _waveterm_si_blocked`，$TMUX 非空时跳过 OSC/钩子 |

---

## 2. ✅ 采用方案 A（已确认）

**核心思路**：新增专用 meta key `term:tmux:session`，值 = tmux 会话名。block 具备该 key 即视为"远程 tmux 会话"。shell 启动时自动 `tmux new -A -t <会话名>`。

### 2.1 实现路径（已验证可行）

```
改动面（集中在 5 处）：
1. 后端：定义 meta key 常量
   - pkg/waveobj/metaconsts.go  ─ 新增 MetaKey_TermTmuxSession = "term:tmux:session"
   - pkg/waveobj/wtypemeta.go   ─ 新增字段 TermTmuxSession

2. 后端：shell 启动注入自动 attach（核心）
   - pkg/blockcontroller/blockcontroller.go makeSwapToken()
   - 检测 blockMeta 的 term:tmux:session
   - 若存在 → 在 token.ScriptText 末尾（或开头）追加：
        \n[ -n "$WAVETERM_CONN" ] && command -v tmux >/dev/null 2>&1 && exec tmux new -A -t "<会话名>"\n
     （内联脚本形式，remote SSH block 才发起，tmux 存在才执行，用 exec 替换 shell）

3. 前端：提供设置入口（可选但推荐）
   - 右键菜单 / block 属性里加「Tmux 会话名」输入项，写回 meta

4. 前端 wsh 命令支持（如需 setmeta）
   - wsh setmeta 已存在，可直接用 wsh setmeta 'term:tmux:session=omlx-11335'

5. schema 更新（如配置校验需要）
   - schema/settings.json 或 blockmeta schema
```

### 2.2 注入代码设计（伪码，写在 makeSwapToken 附近/或新函数）

```go
// 在 makeSwapToken 构造 token 后：
tmuxSession := blockMeta.GetString(waveobj.MetaKey_TermTmuxSession, "")
if tmuxSession != "" && !conncontroller.IsLocalConnName(remoteName) {
    guard := `[ "$TMUX" != "" ] || ! command -v tmux >/dev/null 2>&1 || `
    attach := fmt.Sprintf(`exec tmux new -A -t "%s"`, tmuxSession)
    token.ScriptText += "\n" + guard + attach + "\n"
}
```

> 设计要点：
> - `exec`：attach 成功后 shell 进程被 tmux 替换，避免残留一层空 shell。
> - `[ "$TMUX" != "" ] ||`：已在 tmux 里则跳过，防止嵌套。
> - `command -v tmux`：远端未装 tmux 时优雅跳过，不影响普通 Shell。
> - `!conncontroller.IsLocalConnName(remoteName)`：仅远程 SSH block 生效（也有 `WAVETERM_CONN` 兜底）。
> - 会话名做**引号包裹 + 转义**，避免特殊字符。

### 2.3 备选（环境变量注入 + 用户自写 init script）

若不希望自动注入 tmux 命令，改为在 swap token 里注入环境变量：
```go
token.Env["WAVE_TMUX_SESSION"] = tmuxSession  // term:tmux:session 的值
```
用户在自己的 `cmd:initscript.zsh` 里写 `tmux new -A -t "$WAVE_TMUX_SESSION"`。
**优点**：对已有 init script 用户更灵活；**缺点**：需用户自行写脚本，不够"即开即用"。

> 二者可融合：默认自动注入（2.2），若用户已配置 cmd:initscript 则只注入环境变量（2.3），由用户脚本接管。

---

## 3. 验收标准（草案）

1. 打开带 `term:tmux:session` 的远程 block → 自动进入对应 tmux 会话（无需手动命令）。
2. 会话不存在时自动创建；存在时 attach 到老的会话。
3. 修改 meta 的会话名 → 下次打开 attach 到新会话名。
4. 普通 block（无该 meta）行为不变；本地 block 不触发。
5. 远端无 tmux 时优雅回退为普通 shell，不报错。
6. 已在 tmux 内时不会嵌套新 tmux。

---

## 4. 风险与注意事项

- **TMUX 环境继承**：`tmux new -A` 需要 tmux server 已由同一本地用户启动；`tmux new -A -t` 复用 server 时窗口布局可能共享。需验证多次 attach 是否影响独立窗口。
- **exec 替换 shell 后**：Wave 的 shell integration（OSC/precmd）在该 pane 内不可用，属已知取舍（手动 attach 亦然）。
- **会话名安全**：会话名来自用户 meta，需转义防注入（`tmux -L` 也考虑）。
- **SSH ProxyJump / 非默认 shell**：验证 init 脚本在 zsh(ZDOTDIR) / bash(--rcfile) 分支下都能注入。
- **与 durable session 的关系**：本功能独立于 durable，二者不冲突；但若 block 同时开了 durable 与 term:tmux:session，行为需定义（建议 durable 时忽略 tmux，或反之）。

