[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-examples/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-examples/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-examples)](https://pkg.go.dev/github.com/yylego/kratos-examples)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-examples/main.svg)](https://coveralls.io/github/yylego/kratos-examples?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-examples.svg)](https://github.com/yylego/kratos-examples/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-examples)](https://goreportcard.com/report/github.com/yylego/kratos-examples)

# kratos-examples

基于 Go-Kratos 框架构建的演示项目

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 项目简介

**kratos-examples** 是使用 [Go-Kratos](https://go-kratos.dev) 框架构建微服务的最佳实践参考实现。它的作用是：

- 🎯 **基础项目** - kratos-orz 生态系统中 20+ 专用演示项目的基础模板
- 🛠️ **工具链集成示例** - 展示 kratos-orz 开发工具的实践应用
- 📚 **学习资源** - 完整的微服务架构，遵循 Kratos 规范
- ⚡ **快速开发** - 通过 make orz 等魔法命令实现 proto 和代码自动同步

## 关于 Go-Kratos

[Go-Kratos](https://go-kratos.dev) 是一个简洁高效的微服务框架，它提供：

- 清晰的架构和关注点分离
- 完善的中间件和插件生态
- 内置 gRPC 和 HTTP 协议支持
- 优秀的文档和活跃的生态

**kratos-examples 在这个坚实的基础上**，添加了增强的工具链和自动化能力，简化开发工作流程。

## 示例项目

以下项目均基于 kratos-examples 进行 fork，分别展示不同的功能特性和集成方案，涵盖认证、配置、定时任务、数据库、日志、链路追踪、前端集成等多个方面：

|    演示     |                      仓库                      |
| :---------: | :--------------------------------------------: |
|     ast     |     https://github.com/kratos-examples/ast     |
| custom-auth | https://github.com/kratos-examples/custom-auth |
| static-auth | https://github.com/kratos-examples/static-auth |
|   config    |   https://github.com/kratos-examples/config    |
|    cron     |    https://github.com/kratos-examples/cron     |
|     ebz     |     https://github.com/kratos-examples/ebz     |
|    cobra    |    https://github.com/kratos-examples/cobra    |
|    gorm     |    https://github.com/kratos-examples/gorm     |
|  gormzhcn   |  https://github.com/kratos-examples/gormzhcn   |
|    cors     |    https://github.com/kratos-examples/cors     |
|    i18n     |    https://github.com/kratos-examples/i18n     |
|    nacos    |    https://github.com/kratos-examples/nacos    |
| rate-limit  | https://github.com/kratos-examples/rate-limit  |
|   swaggo    |   https://github.com/kratos-examples/swaggo    |
|    trace    |    https://github.com/kratos-examples/trace    |
|    test     |    https://github.com/kratos-examples/test     |
|    vue3     |    https://github.com/kratos-examples/vue3     |
|  vue3zhcn   |  https://github.com/kratos-examples/vue3zhcn   |
|    wire     |    https://github.com/kratos-examples/wire     |
|     zap     |     https://github.com/kratos-examples/zap     |
|    zapzh    |    https://github.com/kratos-examples/zapzh    |
|   migrate   |   https://github.com/kratos-examples/migrate   |
|    ping     |    https://github.com/kratos-examples/ping     |
| supervisord | https://github.com/kratos-examples/supervisord |

## 核心功能

### 🛠️ 推荐工具链

kratos-orz 生态提供一组专注的小工具。前几个面向 Kratos 与 proto 工作流，其余是通用的 Go 小工具。按需取用即可——每条命令各自独立（不编号，方便日后增减），多数 IDE 会在代码块上显示运行按钮，点一下即可安装，每个仓库链接里有完整文档。

#### `orzkratos-add-proto`

在 api 目录下生成一个新的 proto 文件，新接口从准备好的骨架起步。

仓库：https://github.com/yylego/kratos-orz

```bash
go install github.com/yylego/kratos-orz/cmd/orzkratos-add-proto@latest
```

#### `orzkratos-srv-proto`

让 service 代码与 proto 契约保持同步：补全方法框架、隐藏已删方法、按 proto 顺序重排。

仓库：https://github.com/yylego/kratos-orz

```bash
go install github.com/yylego/kratos-orz/cmd/orzkratos-srv-proto@latest
```

#### `protoc-gen-orzkratos-errors`

从 proto 枚举生成 Go 错误码辅助函数，带状态码和错误嵌套。

仓库：https://github.com/yylego/kratos-errgen

```bash
go install github.com/yylego/kratos-errgen/cmd/protoc-gen-orzkratos-errors@latest
```

#### `wirekratos`

在 Kratos 工作区模式下运行 Wire 依赖注入，带框架感知标识。

仓库：https://github.com/yylego/kratos-wire

```bash
go install github.com/yylego/kratos-wire/cmd/wirekratos@latest
```

#### `clang-format-batch`

一次性批量格式化 proto 和 cpp 源码。

仓库：https://github.com/yylego/clang-format

```bash
go install github.com/yylego/clang-format/cmd/clang-format-batch@latest
```

#### `depbump`

一次性升级模块依赖，需要时可覆盖整个工作区。

仓库：https://github.com/yylego/depbump

```bash
go install github.com/yylego/depbump/cmd/depbump@latest
```

#### `go-lint`

封装 golangci-lint 并自动格式化，跨模块运行格式化与静态检查。

仓库：https://github.com/yylego/go-lint

```bash
go install github.com/yylego/go-lint/cmd/go-lint@latest
```

#### `go-commit`

自动化 git 提交，内置 Go 代码格式化。

仓库：https://github.com/yylego/go-commit

```bash
go install github.com/yylego/go-commit/cmd/go-commit@latest
```

#### `tago`

管理 git 标签并语义化自增，也处理带路径前缀的子模块标签。

仓库：https://github.com/yylego/tago

```bash
go install github.com/yylego/tago/cmd/tago@latest
```

### ⚡ 魔法命令：`make orz`

最强大的功能 - proto 文件与服务代码之间的自动同步：

```bash
make orz
```

**它的作用：**

- ✅ proto 中新增方法 → 自动生成服务方法框架
- ✅ 删除的方法 → 转换成非导出函数（保留逻辑）
- ✅ 重排序方法 → 自动调整服务代码顺序以匹配 proto

**工作流示例：**

1. 在 proto 文件中添加 `CreateArticle` 方法
2. 运行 `make orz`
3. 服务自动生成 `CreateArticle` 方法框架
4. 实现业务逻辑

## 升级与同步思路

本项目是多模块仓库：一个根模块加两个子模块（`demo1kratos`、`demo2kratos`）。根模块引用子模块，所以发布分两轮——先给子模块打标签（带路径前缀，如 `demo1kratos/v0.0.X`），再把根模块升级到子模块的新版本并给根模块打标签（`v0.0.X`）。

下游 fork 项目各自专注一个功能（trace、gorm、zap、i18n 等），按固定节奏从本上游同步：合并上游 main、升级依赖、重新生成代码，然后测试与代码检查。Fork 不会合并回上游——每个都作为独立示例长期存在，持续拉取上游变更以保持与最新 Kratos 版本对齐。

上面的工具正好对应各个环节：`depbump` 管依赖、`tago` 管标签、`go-lint` 管格式化与检查。具体顺序最好一步一步来、不写死成脚本，这样流程能随具体情况灵活调整。

### 代码变更

[changes/](./changes) 包含记录代码差异的 markdown 文件：

- [changes/demo1.md](./changes/demo1.md) - Demo1 相对源项目的变更
- [changes/demo2.md](./changes/demo2.md) - Demo2 相对源项目的变更
- [changes/aside.md](./changes/aside.md) - 附属模块和兄弟项目

这些文件通过测试自动生成：

```bash
go test -v -run TestGenerateDemo1Changes # 生成 demo1.md
go test -v -run TestGenerateDemo2Changes # 生成 demo2.md
go test -v -run TestGenerateAsideChanges # 生成 aside.md
```

**在源项目中：** 文件显示 `✅ NO CHANGES`

**在 fork 项目中：** 文件显示代码差异并带语法高亮，简单直观地在 GitHub 上查看定制化的改动。

## 项目结构

### 内置演示

提供两个内置演示项目，展示各种功能的使用：

- [demo1kratos](./demo1kratos) - Student CRUD 微服务（Kratos 简单示例）
- [demo2kratos](./demo2kratos) - Article CRUD 微服务（高级功能和集成）

两个演示都遵循标准的 Kratos 项目结构，采用 proto-first 设计、Wire 依赖注入、gRPC/HTTP 双协议端点。

### 参数校验

项目演示了双层校验模式：

- **Service 层** - 返回 `ErrorBadParam`（HTTP 400）提示客户端输入无效，给予可操作的反馈
- **Biz 层** - 使用 `must` 断言（失败时 panic）作为安全兜底。由于 Service 层已经做过校验，Biz 层只是冗余检查，因此采用简单的断言而非返回详细错误。这也确保了 Biz 模块自身的安全性——即使其他调用方在新的场景中复用该模块时忘记校验参数，断言也能及时捕获问题。相比在代码中到处散布防御性检查，在入口处用断言直接拒绝非法参数，让后续的业务逻辑保持简洁和确定性

### 数据层

两个演示通过 GORM 使用同一个 PostgreSQL 数据库。

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/yylego/kratos-examples.svg?variant=adaptive)](https://starchart.cc/yylego/kratos-examples)
