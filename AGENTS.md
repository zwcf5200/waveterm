# Wave Terminal - 代理开发指南

本文档为 AI 代理（如 OpenCode Agents）提供 Wave Terminal 代码库的开发规范和指南。

## 构建命令

### 项目初始化
```bash
task init  # 安装所有依赖（npm + go）
```

### 开发模式
```bash
task dev           # 运行 Electron 应用（支持热重载）
task start         # 运行 Electron 应用（不启用热重载）
task quickdev      # 快速开发模式（仅 arm64 macOS）
task winquickdev   # 快速开发模式（仅 Windows amd64）
```

### 构建命令
```bash
task build:dev           # 开发模式构建前端
task build:prod          # 生产模式构建前端
task build:backend       # 构建后端（wavesrv + wsh）
task build:schema        # 构建配置 schema
task generate            # 生成 TypeScript 绑定
task package             # 打包应用（生产构建 + electron-builder）
```

### 测试命令
```bash
npm run test            # 运行所有测试
npm run coverage        # 运行测试并生成覆盖率报告
# 运行单个测试文件
npx vitest run <test-file>              # 运行特定测试文件
npx vitest run <test-file> --reporter=verbose  # 详细输出
npx vitest run -t "<test-name>"         # 运行匹配名称的测试
```

### 代码检查
```bash
task check:ts           # TypeScript 类型检查
go vet ./...            # Go 代码检查
golangci-lint run       # Go lint 检查
```

## 代码风格规范

### TypeScript/React

#### 导入顺序
1. React 和相关库
2. 第三方库
3. 项目内部模块（使用路径别名）
4. 相对路径导入
5. 类型导入
6. 样式文件

```typescript
import { useEffect, useState } from "react";
import clsx from "clsx";
import { useAtomValue } from "jotai";

import { GlobalModel } from "@/app/store/global-model";
import { Workspace } from "@/app/workspace/workspace";

import { BlockFrame } from "./blockframe";
import "./block.scss";
```

#### 路径别名
- `@/app/*` - frontend/app/*
- `@/builder/*` - frontend/builder/*
- `@/util/*` - frontend/util/*
- `@/layout/*` - frontend/layout/*
- `@/store/*` - frontend/app/store/*
- `@/view/*` - frontend/app/view/*
- `@/element/*` - frontend/app/element/*
- `@/shadcn/*` - frontend/app/shadcn/*

#### 格式化规则
- 使用 Prettier：printWidth=120，trailingComma="es5"，useTabs=false
- 2 空格缩进
- 单行最大长度：120 字符
- 强制尾随逗号（ES5 风格）

#### 命名约定
- **组件**：PascalCase（如 `BlockFrame`, `Workspace`）
- **Hooks**：camelCase，以 `use` 开头（如 `useAtomValue`, `useDimensions`）
- **常量**：UPPER_SNAKE_CASE（如 `MAX_RETRIES`）
- **类型/接口**：PascalCase（如 `BlockProps`, `TabModel`）
- **函数**：camelCase（如 `getBlockComponentModel`, `registerBlockComponentModel`）
- **私有变量**：以下划线开头（如 `_internalState`）

#### 错误处理
- 永远不要使用空 catch 块
- 使用 ErrorBoundary 包裹关键组件
- 优先使用 async/await 而非 Promise 链
- 明确错误类型，使用 `unknown` 而非 `any`

```typescript
// 正确
try {
    await someAsyncOperation();
} catch (error) {
    console.error("Operation failed:", error);
    handleError(error);
}

// 错误
catch (e) {}
```

#### 类型安全
- **禁止**使用 `as any`、`@ts-ignore`、`@ts-expect-error`
- 为组件 props 定义明确的接口
- 使用 JSDoc 注释复杂的类型

#### 注释和文档
- 版权声明：每文件开头包含版权声明
- 复杂逻辑必须添加中文注释
- 导出的公共函数和组件必须添加 JSDoc

```typescript
// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

/**
 * 获取块组件模型
 * @param blockId - 块的唯一标识符
 * @returns 块组件模型实例
 */
function getBlockComponentModel(blockId: string): BlockComponentModel {
    // 实现逻辑
}
```

### Go 代码

#### 命名约定
- **包名**：小写单数（如 `config`, `wconfig`）
- **接口**：PascalCase，以 `er` 结尾（如 `Reader`, `Writer`）
- **导出函数**：PascalCase（如 `GetConfig`, `SaveConfig`）
- **私有函数**：小写（如 `parseConfig`, `validateInput`）

#### 格式化
- 使用 `gofmt` 自动格式化
- 运行 `go fmt ./...` 在提交前

#### 错误处理
- 总是检查错误
- 使用 `fmt.Errorf` 或自定义错误类型
- 不要忽略返回的 error 值

```go
// 正确
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}

// 错误
result, _ := someFunction()
```

## 项目结构

```
waveterm/
├── frontend/          # 前端 React 应用
│   └── app/
│       ├── aipanel/   # AI 面板组件
│       ├── block/     # 块组件（终端、预览等）
│       ├── element/   # UI 元素
│       ├── store/     # 状态管理（Jotai）
│       ├── view/      # 视图模型
│       └── workspace/ # 工作区管理
├── pkg/              # Go 后端包
├── cmd/              # Go 命令行工具
├── emain/            # Electron 主进程
├── tsunami/          # Web 组件框架
└── docs/             # 文档站点
```

## 关键技术栈

### 前端
- **框架**：React 19 + TypeScript
- **构建工具**：Vite + electron-vite
- **状态管理**：Jotai
- **UI 组件**：自定义组件 + Tailwind CSS
- **拖放**：react-dnd
- **终端**：xterm.js
- **编辑器**：Monaco Editor
- **测试**：Vitest

### 后端
- **语言**：Go 1.x
- **数据库**：SQLite
- **WebSocket**：用于实时通信

## 开发注意事项

1. **语言要求**：所有文档和代码注释必须使用中文
2. **用户沟通**：始终使用中文与用户沟通
3. **热重载**：开发时使用 `task dev` 以启用 HMR
4. **类型检查**：在提交前运行 `task check:ts` 确保无类型错误
5. **测试覆盖**：添加新功能时必须编写测试
6. **兼容性**：支持 macOS、Linux 和 Windows（arm64 和 amd64）

## 常见问题

- **开发服务器启动慢**：先运行 `task build:backend`，再使用 `task quickdev`
- **类型错误**：检查 `tsconfig.json` 中的路径别名配置
- **Go 编译失败**：确保已安装 Zig（用于 CGO 静态链接）

## 提交前检查清单

- [ ] 运行 `task check:ts` 通过
- [ ] 运行 `go fmt ./...` 和 `go vet ./...` 通过
- [ ] 新功能包含测试并通过
- [ ] 代码通过 ESLint 检查
- [ ] 使用 Prettier 格式化代码
- [ ] 所有文档和注释使用中文
- [ ] 不包含 `as any`、`@ts-ignore` 等类型错误抑制
