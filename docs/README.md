# Arcade 文档中心

欢迎来到 Arcade CI/CD 平台文档中心！

## 📚 文档导航

### 快速开始

- **[快速开始指南](./QUICKSTART.md)** - 了解如何快速上手 Arcade 平台
- **[标签示例](./LABEL_EXAMPLES.md)** - Agent 标签系统使用示例

### 插件系统

#### 新手入门
- **[插件快速开始](./PLUGIN_QUICKSTART.md)** ⭐ 推荐
  - 5分钟快速体验
  - 简单示例
  - 常用命令

#### 深入学习
- **[插件开发指南](./PLUGIN_DEVELOPMENT.md)** 📖 必读
  - 完整的开发教程
  - 6种插件类型详解
  - 实战示例
  - 最佳实践
  - 常见问题

#### 参考资料
- **[插件快速参考](./PLUGIN_REFERENCE.md)** 🔖 速查
  - 代码片段
  - 常用命令
  - 接口速查表
  - 配置示例

#### 高级主题
- **[插件自动加载](./PLUGIN_AUTO_LOAD.md)** 🚀 进阶
  - 自动监控原理
  - 热重载机制
  - API 参考
  - 故障排除

## 🎯 根据需求查找文档

### 我想...

#### ...快速体验插件系统
→ [插件快速开始](./PLUGIN_QUICKSTART.md)

#### ...开发一个新插件
→ [插件开发指南](./PLUGIN_DEVELOPMENT.md)

#### ...查找代码示例
→ [插件快速参考](./PLUGIN_REFERENCE.md)

#### ...了解自动加载功能
→ [插件自动加载](./PLUGIN_AUTO_LOAD.md)

#### ...使用 Agent 标签系统
→ [标签示例](./LABEL_EXAMPLES.md)

#### ...了解整体架构
→ [快速开始指南](./QUICKSTART.md)

## 📖 文档内容概览

### QUICKSTART.md
- Arcade 平台概述
- 架构说明
- API 服务介绍
- Label 系统详解

### LABEL_EXAMPLES.md
- Label 使用场景
- LabelSelector 示例
- 最佳实践

### PLUGIN_QUICKSTART.md
- 5分钟快速体验
- 基础用法
- 开发第一个插件
- 常用命令

### PLUGIN_DEVELOPMENT.md
**最完整的插件开发教程**
- 插件系统概述（原理图）
- 6种插件类型详解
- 开发环境准备
- 详细开发步骤
- 插件接口详解
- 2个完整实战示例
  - 钉钉通知插件
  - Docker 构建插件
- 编译和调试技巧
- 8个最佳实践
- 常见问题解答
- 插件发布指南

### PLUGIN_REFERENCE.md
**快速参考手册**
- 插件类型速查表
- 最小/完整插件模板
- 常用代码片段
  - 配置解析
  - HTTP 客户端
  - 数据库连接
  - 后台任务
  - 并发安全
  - 错误处理
- 常用命令
- 配置示例
- 接口实现速查
- 常见错误表
- 性能优化技巧
- 安全建议

### PLUGIN_AUTO_LOAD.md
**自动加载详细说明**
- 功能特性
- 使用方法
- 监控原理
- 防抖机制
- 注意事项
- 使用场景
- API 参考
- 故障排除

## 🔧 实用工具

### 快速命令

```bash
# 编译所有插件
make plugins

# 运行插件演示
make example-autoload

# 查看所有命令
make help
```

### 示例代码位置

- **插件示例**: `pkg/plugins/`
- **演示程序**: `examples/plugin_autowatch/`
- **测试脚本**: `scripts/test_plugin_autoload.sh`

## 📝 文档版本

- **版本**: 1.0
- **最后更新**: 2025-01-08
- **适用于**: Arcade v1.x

## 🤝 贡献文档

发现文档问题或想要改进？

1. 提交 Issue 说明问题
2. 提交 Pull Request 改进文档
3. 联系维护者

## 📞 获取帮助

- 查看相关文档
- 查看示例代码
- 提交 Issue
- 联系开发团队

---

**提示**: 文档之间互相链接，你可以从任何一个文档跳转到相关内容。

祝使用愉快！🎉

