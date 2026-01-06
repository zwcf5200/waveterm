# Fork 仓库同步指南

本文档说明如何将 Fork 仓库与原仓库（Upstream）保持同步。

## 📋 前置条件

- 已 Fork 原仓库到自己的 GitHub 账号
- 已克隆 Fork 仓库到本地
- 有 Git 和 GitHub CLI（gh）的访问权限

## 🚀 一、配置上游仓库

### 1. 查看当前远程仓库配置

```bash
git remote -v
```

**预期输出：**

```
origin  git@github.com:你的用户名/仓库名.git (fetch)
origin  git@github.com:你的用户名/仓库名.git (push)
```

### 2. 添加上游仓库（Upstream）

```bash
git remote add upstream git@github.com:原仓库所有者/仓库名.git
```

**示例：**

```bash
git remote add upstream git@github.com:wavetermdev/waveterm.git
```

### 3. 验证配置

```bash
git remote -v
```

**预期输出：**

```
origin    git@github.com:zwcf5200/waveterm.git (fetch)
origin    git@github.com:zwcf5200/waveterm.git (push)
upstream  git@github.com:wavetermdev/waveterm.git (fetch)
upstream  git@github.com:wavetermdev/waveterm.git (push)
```

---

## 🔄 二、同步上游代码

### 方法 1：使用命令行（推荐 - 完整控制）

#### 步骤 1：获取上游最新代码

```bash
git fetch upstream
```

#### 步骤 2：查看所有分支

```bash
git branch -r
```

#### 步骤 3：切换到你的本地分支（例如 dev）

```bash
git checkout dev
```

#### 步骤 4：合并上游分支到本地分支

**选项 A：同步上游的 dev 分支**

```bash
git merge upstream/dev
```

**选项 B：同步上游的 main 分支到你的 dev 分支**

```bash
git merge upstream/main
```

**选项 C：使用 rebase 保持提交历史更清晰**

```bash
git rebase upstream/dev
# 或
git rebase upstream/main
```

#### 步骤 5：推送到你的 Fork

```bash
git push origin dev
```

**如果是 rebase 操作，可能需要强制推送：**

```bash
git push origin dev --force-with-lease
```

---

### 方法 2：使用 GitHub CLI（快捷方式）

#### 自动同步 Fork 仓库

```bash
gh repo sync 你的用户名/仓库名
```

**示例：**

```bash
gh repo sync zwcf5200/waveterm
```

这会自动同步原仓库的默认分支到你的 Fork。

#### 同步特定分支

```bash
gh repo sync --branch dev 你的用户名/仓库名
```

**示例：**

```bash
gh repo sync --branch dev zwcf5200/waveterm
```

---

## ⚠️ 三、处理合并冲突

如果合并时出现冲突，Git 会提示：

```
Auto-merging 文件名
CONFLICT (content): Merge conflict in 文件名
Automatic merge failed; fix conflicts and then commit the result.
```

### 解决冲突的步骤

#### 1. 查看冲突文件

```bash
git status
```

#### 2. 打开冲突文件，找到冲突标记

```
<<<<<<< HEAD
你的修改内容
=======
上游的修改内容
>>>>>>> upstream/dev
```

#### 3. 手动编辑文件，解决冲突

保留需要的代码，删除冲突标记。例如：

```javascript
// 解决后的代码
function example() {
  // 合并后的代码
  console.log("merged content");
}
```

#### 4. 标记冲突已解决

```bash
git add <冲突文件名>
```

**或添加所有已解决的文件：**

```bash
git add .
```

#### 5. 完成合并

```bash
git commit
```

#### 6. 推送到你的 Fork

```bash
git push origin dev
```

---

## 📚 四、推荐的工作流

### 日常同步的最佳实践

#### 1. 定期同步上游代码（每次开始新功能前）

```bash
# 获取上游最新代码
git fetch upstream

# 切换到开发分支
git checkout dev

# 合并上游分支
git merge upstream/dev

# 推送到你的 Fork
git push origin dev
```

#### 2. 创建功能分支进行开发

```bash
# 从最新的 dev 分支创建功能分支
git checkout dev
git pull origin dev
git checkout -b feature/your-feature-name

# 开发代码...
git add .
git commit -m "添加你的功能"

# 推送功能分支
git push origin feature/your-feature-name
```

#### 3. 功能完成后，再次同步上游并合并

```bash
# 切换回 dev 分支
git checkout dev

# 同步上游最新代码
git fetch upstream
git merge upstream/dev
git push origin dev

# 在 GitHub 上创建 Pull Request
```

---

## 🛠️ 五、常用命令速查

### 远程仓库管理

```bash
# 查看远程仓库
git remote -v

# 添加上游仓库
git remote add upstream <url>

# 删除远程仓库
git remote remove <name>

# 重命名远程仓库
git remote rename <old> <new>

# 查看远程仓库详细信息
git remote show origin
git remote show upstream
```

### 分支管理

```bash
# 查看本地分支
git branch

# 查看所有分支（本地和远程）
git branch -a

# 查看远程分支
git branch -r

# 创建新分支
git branch <branch-name>

# 切换分支
git checkout <branch-name>

# 创建并切换到新分支
git checkout -b <branch-name>

# 删除本地分支
git branch -d <branch-name>

# 删除远程分支
git push origin --delete <branch-name>
```

### 同步操作

```bash
# 获取远程仓库更新
git fetch origin
git fetch upstream

# 拉取远程更新并合并
git pull origin dev

# 合并分支
git merge <branch-name>

# 变基分支
git rebase <branch-name>

# 推送本地分支到远程
git push origin <branch-name>

# 推送所有分支
git push origin --all

# 强制推送（谨慎使用）
git push origin <branch-name> --force-with-lease
```

---

## 🎯 六、故障排除

### 问题 1：无法推送到 upstream

**错误信息：**

```
ERROR: Permission to 原仓库/仓库名.git denied to 你的用户名.
fatal: Could not read from remote repository.
```

**解决方案：**

- 你不应该推送到 upstream，只能推送到 origin（你的 Fork）
- upstream 只用于拉取，不用于推送

### 问题 2：合并后历史混乱

**解决方案：**

- 使用 `git rebase upstream/dev` 而不是 `git merge`
- 或者使用 `git pull --rebase`

### 问题 3：本地分支与上游分支偏离

**查看情况：**

```bash
git log --oneline --graph --all
```

**解决方案：**

```bash
# 重置到上游分支（谨慎使用，会丢失本地修改）
git reset --hard upstream/dev

# 或者保留本地修改作为备份
git stash
git reset --hard upstream/dev
git stash pop
```

### 问题 4：无法删除分支（分支被保护）

**解决方案：**

- 在 GitHub 仓库设置中取消分支保护
- 或在本地使用 `-D` 强制删除

```bash
git branch -D <branch-name>
```

---

## 📖 七、相关资源

- [GitHub 官方文档：同步 Fork](https://docs.github.com/zh/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork)
- [Git 官方文档：远程仓库](https://git-scm.com/book/zh/v2/Git-%E5%88%86%E5%B8%83%E5%BC%8F%E5%B7%A5%E4%BD%9C-%E8%BF%9C%E7%A8%8B%E4%BB%93%E5%BA%93)
- [GitHub CLI 文档](https://cli.github.com/manual/)

---

## 💡 八、最佳实践总结

1. **定期同步**：每次开始新功能前，先同步上游最新代码
2. **使用功能分支**：不要直接在主分支上开发
3. **及时解决冲突**：冲突越早解决越容易
4. **保持提交历史清晰**：使用 rebase 或规范的提交信息
5. **备份重要修改**：在重置或强制操作前，使用 `git stash` 备份
6. **团队协作**：同步后及时通知团队成员，避免重复工作

---

## 🔍 九、当前项目配置示例

```
origin    git@github.com:zwcf5200/waveterm.git (fetch)  # 你的 Fork
origin    git@github.com:zwcf5200/waveterm.git (push)   # 你的 Fork
upstream  git@github.com:wavetermdev/waveterm.git (fetch)  # 原仓库
upstream  git@github.com:wavetermdev/waveterm.git (push)   # 原仓库（只读）
```

**日常同步命令：**

```bash
git fetch upstream
git checkout dev
git merge upstream/dev
git push origin dev
```

---

**文档版本：** 1.0
**最后更新：** 2026-01-06
**维护者：** Wave Terminal 开发团队
