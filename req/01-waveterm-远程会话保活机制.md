# Wave Terminal 远程会话保活机制分析

> 需求讨论文档 #1
> 目标：弄清源项目（waveterm）远程 SSH 会话的保活机制，评估「能否改用 tmux 替代」

---

## 1. 结论速览

- **waveterm 自己实现了一套「类 tmux」的 durable session（0.14 引入）**，没有使用 tmux。
- 远程侧跑一个 **独立的 job manager 守护进程**（就是 wsh 本身的 `jobmanager` 子命令），通过 `setsid()` 脱离 SSH 会话进程组实现保活，持久监听 Unix domain socket。
- 保活 = **「远程 shell 进程独立于 SSH 连接存活」 + 「本地 wavesrv 与 job manager 之间的 RPC 重连」** 两层。
- **无法简单用 tmux 替代**：waveterm 需要的是 shell 之外的第二层持久状态（滚动缓冲、输入序列、RPC 认证、block 生命周期绑定），tmux 只解决「进程存活」这一层，不解决 Wave 与远程 shell 之间的可控 RPC 通道。

---

## 2. Durable Session 的完整链路

### 2.1 分层架构

```
┌─────────────────────────────────────────────┐
│ 本地 wavesrv（wave server）                   │
│  ├─ DurableShellController（durableshellcontroller.go）│
│  ├─ jobcontroller（job 生命周期 / 重连）        │
│  └─ wshutil.Router（RPC 路由）                 │
└─────────────────────────────────────────────┘
              │  SSH 连接（到远程 connserver）
              ▼
┌─────────────────────────────────────────────┐
│ 远程 SSH 会话                                │
│  ├─ wsh connserver（RPC server，随 SSH 存活）  │
│  └─ job manager（wsh jobmanager，setsid 保活） │
│        │  Unix socket (/tmp/waveterm-job-*)  │
│        ▼                                     │
│  ┌─────────────────────────────┐            │
│  │ 真·shell 进程（bash/zsh）     │            │
│  └─────────────────────────────┘            │
└─────────────────────────────────────────────┘
```

### 2.2 关键代码事实

| 层级 | 文件 | 作用 |
|------|------|------|
| Block 控制器 | `pkg/blockcontroller/durableshellcontroller.go` | 每个 terminal block 绑定一个 JobId；Start/Stop/重连 |
| Job 启动 | `pkg/shellexec/shellexec.go:472 StartRemoteShellJob` | 确定 shell 路径/参数，调 `jobcontroller.StartJob` |
| Job 控制器 | `pkg/jobcontroller/jobcontroller.go` | Job 生命周期、重连（`ReconnectJob`）、状态机 |
| 远程启动 | `pkg/wshrpc/wshremote/wshremote_job.go:133 RemoteStartJobCommand` | 在远程执行 `wsh jobmanager --jobid <id> --clientid <id>` |
| **保活核心** | `pkg/jobmanager/jobmanager_unix.go daemonize` | **`unix.Setsid()` 脱离会话 + /dev/null 重定向 stdio + 忽略 SIGHUP** |
| 重连核心 | `pkg/jobmanager/mainserverconn.go` + `jobcontroller.go:doReconnectJob` | Job manager 通过 RPC 认证后，本地 wavesrv 拉取并重建立 stream |

### 2.3 保活着眼点拆解

**A. 进程保活（跨 SSH 断线存活）—— 由 job manager 守护**

`daemonize()` 是核心（`jobmanager_unix.go:20`）：
- `unix.Setsid()`：创建新会话，**脱离 SSH 终端的控制终端与进程组**。SSH 连接断开时，会话收 SIGHUP 的是 SSH 会话进程组，而 job manager 已处于新会话，不受影响。
- `unix.Dup2(/dev/null, stdin)`：断开终端 I/O。
- `signal.Ignore(SIGHUP)`：即使收到 SIGHUP 也忽略。
- 还有一个 keepalive goroutine（`jobmanager.go:82`）：每小时 ticker 打日志，让 job manager 保持活动（防止被当作 idle 清理）。

**B. 通信通道保活（SSH 断开后如何续上）—— RPC 重连**

- job manager 在远程监听 **Unix socket**（`wavebase.GetRemoteJobSocketPath`），不依赖 TCP 端口（文档明说"no open ports"）。
- 本地 wavesrv 侧通过 `wshutil.MakeJobRouteId(jobId)` 维护 job 路由。
- SSH 断开时，job 路由 down；**SSH 恢复后**，`jobcontroller.doReconnectJob` 发送 `RemoteReconnectToJobManagerCommand`（携带 JobAuthToken 双向认证）→ 重新连上远程 socket → `restartStreaming` 恢复输出流。

**C. 输入/输出一致性保活**

- 输入走带序号队列：`InputSessionId + SeqNum`（`SendInput`, `jobcontroller.go`），断线期间允许缓冲/追上。
- 输出端 job manager 用 `streammanager` 缓冲，重连后同步滚动缓冲（术语 scrollback）。

### 2.4 状态机（文档确认）

SSH 默认**不启用** durable（opt-in，`term:durable: true`），可全局/连接级/block 级配置。状态：Attached（连接）→ Detached（断开，job 继续跑）→ Awaiting/Starting/Ended。

---

## 3. 关键结论：为什么 job manager 要「自己写」而不用 tmux

官方文档原话（`durable-sessions.mdx:15`）：*"similar to tmux or screen but built directly into Wave"*，并明确说 no open ports、all over SSH。

waveterm 自己写 job manager 而非用 tmux 的**决定性原因**：

1. **需要自己的 RPC 传输层**：Wave 前端要做跨平台的终端同步、滚动缓冲持久化、输入序列补齐、block↔job 生命周期绑定。tmux 不提供这些。
2. **身份认证**：job 与 block、本地 wavesrv、用户会话绑定，需 JWT/JobAuthToken 双向认证。tmux 无法承载。
3. **跨平台**：job manager 是 Go 的 wsh 二进制，Linux/macOS/Windows 统一；tmux 在 Windows 上不可用。
4. **依赖可控**：不引入 tmux 这个外部运行时依赖，避免目标主机未装 tmux / 版本不一致 / 权限问题。

---

## 4. 对「能否用 tmux 替代」的回答

**结论：作为「技术栈替换」不可行**，但可作为**形态参考/混合方案**。

- ❌ 替换 job manager 整体：不现实——tmux 只解决"shell 进程持久"，不解决 Wave 的 RPC 通道、scrollback 同步、认证、block 绑定。
- ✅ 若需求只是「**命令行/脚本在不同 shell 窗口间共享同一持久会话**」（轻量场景），tmux 是现成可用的成熟方案（tmux attach/detach、共享 pane、脚本控制）。
- ⚠️ **混合思路（如果场景允许）**：默认仍用 Wave 自带 durable session（推荐，零额外依赖）；若用户明确要在服务器上已有 tmux 会话中操作，可让 Wave 的 shell startup 在检测到 tmux 时友好共存（而非替换核心保活）。

---

## 5. 决策问题（给需求定义用）

若要做「远程会话保活」需求，先厘清目标：

- Q1：保活目标是「进程不死」（关掉 Wave 后命令继续跑）→ Wave durable session 已实现。
- Q2：还是「多窗口共享同一个 shell」（类似 tmux 的 attach/shared session）→ Wave 目前是 1 block = 1 job = 1 shell，**不支持多窗口 attach 同一会话**，这可能才是开放点。
- Q3：是否需要「断线后从终端滚动缓冲无缝续接」→ Wave 已实现（stream 重连 + scrollback）。

> 如果真正的痛点是 Q2（会话共享/多终端 attach），那么 **tmux 反而是一个合理的方向**，因为 Wave 自身的 durable 机制默认 1:1，而 tmux 天生支持 N:1。
