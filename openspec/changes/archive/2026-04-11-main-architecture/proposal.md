## Why

项目缺乏一个统一的、长期的主规格文档来描述整体架构设计。随着功能迭代，各模块的设计决策分散在代码注释、临时文档和历史变更中，导致新成员难以理解系统全貌，也增加了维护成本。

建立主规格文档的目标是：
- 提供清晰的架构总览，帮助开发者快速理解系统设计
- 作为设计评审的参考基准，确保新功能符合架构原则
- 记录关键架构决策，便于未来追溯和演进

## What Changes

- 创建主规格文档 `specs/architecture/spec.md`，包含：
  - 系统架构总览（分层结构、核心组件）
  - 技术栈与依赖说明
  - 目录结构与模块划分
  - 核心组件交互关系
  - 设计原则与约束

## Capabilities

### New Capabilities
- `architecture`: 项目整体架构设计文档，包含分层架构、模块划分、技术栈、目录结构、核心组件交互等

### Modified Capabilities
- none

## Impact

Affected files:
- 新增：`specs/architecture/spec.md` — 主架构规格文档

Affected behavior:
- 无运行时行为变更
- 作为设计评审和新功能开发的架构参考基准
