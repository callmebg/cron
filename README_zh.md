# Go Cron - 原生定时任务调度器

[![版本](https://img.shields.io/badge/版本-v0.2.0--beta-blue.svg)](https://github.com/callmebg/cron)
[![Go版本](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![测试覆盖率](https://img.shields.io/badge/覆盖率-75.4%25-green.svg)](#测试覆盖率)
[![许可证](https://img.shields.io/badge/许可证-MIT-green.svg)](LICENSE)

一个轻量级、高效的Go应用程序定时任务调度器，完全基于Go标准库构建。该库提供强大的任务调度功能和全面的监控能力，同时保持零外部依赖。

[English](README.md) | 中文

## ✨ 特性

- **纯Go标准库**：无外部依赖，充分利用Go的原生能力
- **标准Cron语法**：完全支持5字段和6字段cron表达式
- **并发执行**：使用goroutines和工作池高效执行任务
- **内置监控**：全面的指标、执行跟踪和HTTP端点
- **优雅关闭**：基于context的清理终止
- **线程安全**：支持多个goroutines并发使用，具备强大同步机制
- **灵活配置**：可配置任务超时、重试和错误处理
- **内存高效**：基于优先队列调度，最小内存占用
- **生产就绪**：全面测试覆盖率(75.4%)，包含集成和基准测试

## 🚀 快速开始

### 安装

```bash
go get github.com/callmebg/cron
```

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // 创建新的调度器
    scheduler := cron.New()

    // 添加每分钟执行的任务
    err := scheduler.AddJob("0 * * * * *", func() {
        fmt.Println("任务执行时间:", time.Now().Format("2006-01-02 15:04:05"))
    })
    if err != nil {
        log.Fatal("添加任务失败:", err)
    }

    // 启动调度器
    ctx := context.Background()
    scheduler.Start(ctx)

    // 运行5分钟
    time.Sleep(5 * time.Minute)

    // 优雅停止调度器
    scheduler.Stop()
    fmt.Println("调度器已停止")
}
```

### 高级配置用法

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // 创建自定义配置
    config := cron.DefaultConfig()
    config.MaxConcurrentJobs = 50
    config.EnableMonitoring = true
    config.MonitoringPort = 8080
    
    // 使用自定义配置创建调度器
    scheduler := cron.NewWithConfig(config)

    // 添加带错误处理和重试配置的任务
    jobConfig := cron.JobConfig{
        MaxRetries:    3,
        RetryInterval: 30 * time.Second,
        Timeout:       5 * time.Minute,
    }

    err := scheduler.AddJobWithConfig("备份任务", "0 0 2 * * *", jobConfig, func() {
        // 模拟备份操作
        log.Println("运行备份任务...")
        time.Sleep(2 * time.Second)
        log.Println("备份成功完成")
    })
    if err != nil {
        log.Fatal("添加备份任务失败:", err)
    }

    // 添加带错误处理器的任务
    err = scheduler.AddJobWithErrorHandler("监控任务", "*/30 * * * * *", 
        func() error {
            // 可能失败的任务
            return performHealthCheck()
        },
        func(err error) {
            log.Printf("健康检查失败: %v", err)
        })
    if err != nil {
        log.Fatal("添加监控任务失败:", err)
    }

    // 使用context启动
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    scheduler.Start(ctx)

    // 通过HTTP监控指标（如果启用监控）
    // curl http://localhost:8080/metrics
    
    // 保持运行
    select {}
}

func performHealthCheck() error {
    // 模拟健康检查逻辑
    return nil
}
```

## 📁 项目结构

项目遵循Go最佳实践，组织清晰，易于维护和贡献：

```
cron/
├── cmd/examples/              # 🎯 示例应用程序
│   ├── basic/                 # 基础用法模式
│   ├── advanced/              # 高级配置
│   ├── monitoring/            # 监控集成
│   └── testing/               # 测试模式
├── internal/                  # 🔒 内部包
│   ├── parser/                # Cron表达式解析
│   ├── scheduler/             # 核心调度逻辑
│   ├── monitor/               # 指标和监控
│   ├── types/                 # 内部类型定义
│   └── utils/                 # 工具函数
├── pkg/cron/                  # 📦 公共API
│   ├── cron.go                # 主调度器实现
│   ├── config.go              # 配置管理
│   ├── errors.go              # 错误定义
│   └── types.go               # 公共类型
├── test/                      # 🧪 测试套件
│   ├── integration/           # 集成测试
│   ├── benchmark/             # 性能基准测试
│   └── testdata/              # 测试数据和夹具
├── docs/                      # 📚 文档
├── scripts/                   # 🛠️ 构建和测试脚本
└── .github/workflows/         # 🔄 CI/CD管道
```

## 🎯 综合示例

`cmd/examples/` 目录包含全面的可运行示例，演示cron库的各个方面：

### 📚 可用示例

| 示例 | 描述 | 关键特性 |
|------|------|---------|
| **[basic/](cmd/examples/basic/)** | 基础用法模式 | 简单调度、多种cron格式、优雅关闭 |
| **[advanced/](cmd/examples/advanced/)** | 高级配置 | 错误处理、重试、超时、自定义配置 |
| **[monitoring/](cmd/examples/monitoring/)** | 监控和指标 | HTTP端点、实时指标、性能跟踪 |
| **[testing/](cmd/examples/testing/)** | 测试模式 | 单元测试、模拟、基准测试、测试助手 |

### 🏃 运行示例

```bash
# 基础用法示例
go run cmd/examples/basic/main.go

# 带错误处理的高级特性
go run cmd/examples/advanced/main.go

# 带HTTP端点的监控
go run cmd/examples/monitoring/main.go
# 访问 http://localhost:8080 查看指标仪表板

# 测试模式（包含测试函数和助手）
go run cmd/examples/testing/main.go
```

### 💡 示例重点

**基础示例**：从这里开始理解基本概念
- 多种任务调度模式
- 5字段和6字段cron表达式
- 正确的启动和关闭程序

**高级示例**：生产就绪模式
- 自定义调度器配置
- 任务重试机制和超时
- 综合错误处理
- 工作时间调度
- 实时状态监控

**监控示例**：可观测性和指标
- HTTP监控端点（`/health`、`/metrics`、`/jobs`）
- 实时性能跟踪
- 任务执行统计
- 基于Web的仪表板访问

**测试示例**：质量保证模式
- 单元测试策略
- 模拟时间实现
- 测试助手工具
- 基准测试编写指南

## 📋 Cron表达式格式

本库支持5字段和6字段cron表达式：

### 6字段格式（推荐）:
```
┌─────────────── 秒 (0-59)
│ ┌─────────────── 分钟 (0-59)
│ │ ┌─────────────── 小时 (0-23)
│ │ │ ┌─────────────── 日期 (1-31)
│ │ │ │ ┌─────────────── 月份 (1-12)
│ │ │ │ │ ┌─────────────── 星期 (0-6, 周日=0)
│ │ │ │ │ │
* * * * * *
```

### 5字段格式（传统）:
```
┌─────────────── 分钟 (0-59)
│ ┌─────────────── 小时 (0-23)
│ │ ┌─────────────── 日期 (1-31)
│ │ │ ┌─────────────── 月份 (1-12)
│ │ │ │ ┌─────────────── 星期 (0-6, 周日=0)
│ │ │ │ │
* * * * *
```

### 支持的操作符

| 操作符 | 描述 | 示例 |
|--------|------|------|
| `*` | 任意值 | `* * * * * *` (每秒) |
| `,` | 值列表 | `0,15,30,45 * * * * *` (每15秒) |
| `-` | 范围 | `0-5 * * * * *` (0到5秒) |
| `/` | 步长值 | `*/10 * * * * *` (每10秒) |

### 表达式示例

| 表达式 | 描述 |
|--------|------|
| `0 * * * * *` | 每分钟 |
| `0 0 * * * *` | 每小时 |
| `0 0 0 * * *` | 每天午夜 |
| `0 0 9 * * 1-5` | 工作日上午9点 |
| `0 */15 * * * *` | 每15分钟 |
| `0 0 0 1 * *` | 每月第一天 |
| `*/30 * * * * *` | 每30秒 |

## ⚙️ 配置选项

### 调度器配置

```go
type Config struct {
    Logger            *log.Logger    // 自定义日志实例
    Timezone          *time.Location // 调度评估时区
    MaxConcurrentJobs int            // 最大并发任务执行数
    EnableMonitoring  bool           // 启用HTTP监控端点
    MonitoringPort    int            // 监控HTTP服务器端口
}
```

### 任务配置

```go
type JobConfig struct {
    MaxRetries    int           // 失败时最大重试次数
    RetryInterval time.Duration // 重试尝试间延迟
    Timeout       time.Duration // 每个任务最大执行时间
}
```

### 默认值

```go
// 默认调度器配置
config := cron.DefaultConfig()
// Logger: os.Stdout with "CRON: " 前缀
// Timezone: time.Local
// MaxConcurrentJobs: 100
// EnableMonitoring: false
// MonitoringPort: 8080

// 默认任务配置
jobConfig := cron.DefaultJobConfig()
// MaxRetries: 0 (无重试)
// RetryInterval: 1分钟
// Timeout: 0 (无超时)
```

## 📊 监控和指标

调度器提供全面的监控能力：

### HTTP监控端点

启用监控时，调度器通过HTTP暴露指标：

```bash
# 获取调度器统计信息
curl http://localhost:8080/stats

# 获取详细任务信息
curl http://localhost:8080/jobs

# 获取健康检查
curl http://localhost:8080/health
```

### 可用指标

```go
type Stats struct {
    TotalJobs            int           // 总调度任务数
    RunningJobs          int           // 当前执行任务数
    SuccessfulExecutions int64         // 总成功执行数
    FailedExecutions     int64         // 总失败执行数
    AverageExecutionTime time.Duration // 平均执行时间
    Uptime               time.Duration // 调度器运行时间
}
```

### 编程访问

```go
// 获取调度器统计信息
stats := scheduler.GetStats()
fmt.Printf("总任务数: %d\n", stats.TotalJobs)
fmt.Printf("成功率: %.2f%%\n", 
    float64(stats.SuccessfulExecutions) / 
    float64(stats.SuccessfulExecutions + stats.FailedExecutions) * 100)

// 获取单个任务统计信息
jobStats := scheduler.GetJobStats("备份任务")
fmt.Printf("任务 %s: %d 次运行, %.2f%% 成功率\n", 
    jobStats.Name, jobStats.TotalRuns, jobStats.SuccessRate)
```

## 🚦 错误处理

库提供全面的错误处理机制：

### 内置错误

```go
var (
    ErrInvalidSchedule       = errors.New("无效的cron调度表达式")
    ErrJobNotFound          = errors.New("任务未找到")
    ErrJobAlreadyExists     = errors.New("此名称的任务已存在")
    ErrSchedulerNotStarted  = errors.New("调度器未启动")
    ErrSchedulerAlreadyStarted = errors.New("调度器已启动")
    ErrJobTimeout           = errors.New("任务执行超时")
    ErrMaxRetriesExceeded   = errors.New("超过最大重试次数")
)
```

### 错误处理模式

```go
// 方法1：带错误返回的任务
scheduler.AddJobWithErrorHandler("风险任务", "*/5 * * * * *",
    func() error {
        if err := riskyOperation(); err != nil {
            return fmt.Errorf("操作失败: %w", err)
        }
        return nil
    },
    func(err error) {
        log.Printf("任务失败: %v", err)
        // 发送警报、记录到外部系统等
    })

// 方法2：panic恢复（自动）
scheduler.AddJob("易panic任务", "*/1 * * * * *", func() {
    // 即使这里panic，调度器也会继续运行
    panic("出了点问题")
})

// 方法3：超时处理
jobConfig := cron.JobConfig{
    Timeout: 30 * time.Second, // 30秒后任务将被取消
}
scheduler.AddJobWithConfig("长任务", "0 * * * * *", jobConfig, func() {
    time.Sleep(60 * time.Second) // 这会被取消
})
```

## 🔒 线程安全

调度器完全线程安全，支持：

- **并发任务执行**：多个任务可以同时运行
- **安全任务管理**：调度器运行时添加/移除任务
- **Context取消**：优雅关闭和适当清理
- **资源同步**：内部锁防止竞态条件

```go
// 从多个goroutines调用是安全的
go func() {
    scheduler.AddJob("任务1", "*/1 * * * * *", func() { /* 工作 */ })
}()

go func() {
    scheduler.RemoveJob("旧任务")
}()

go func() {
    stats := scheduler.GetStats() // 线程安全的读访问
}()
```

## 🧪 测试覆盖率

项目在所有组件上保持高测试覆盖率：

| 组件 | 覆盖率 | 类型 |
|------|--------|------|
| 解析器 | 80.8% | 单元 + 集成 |
| 调度器 | 97.3% | 单元 + 集成 + 基准 |
| 队列 | 97.3% | 单元 + 集成 |
| 核心API | 87.3% | 单元 + 集成 + 基准 |
| 类型 | 88.9% | 单元 |
| 工具 | 52.2% | 单元 |
| **总体** | **75.4%** | **所有类型** |

### 运行测试

```bash
# 运行所有测试
go test ./...

# 带覆盖率运行
go test -cover ./...

# 运行集成测试
go test ./test/integration/...

# 运行基准测试
go test -bench=. ./test/benchmark/...

# 使用提供的脚本
./scripts/test.sh        # 全面测试套件
./scripts/benchmark.sh   # 性能基准测试
```

## 📈 性能特征

### 基准测试 (Go 1.19+, Intel i7-12700)

| 操作 | 性能 | 内存 |
|------|------|------|
| Cron解析 | ~3,400 ns/op | 最小分配 |
| 任务调度 | ~1,500 ns/op | < 1KB 每任务 |
| 队列操作 | ~100 ns/op | O(log n) 复杂度 |
| 并发访问 | 线性扩展 | 线程安全 |

### 可扩展性

- **任务容量**：测试支持10,000+并发任务
- **内存使用**：1,000个活跃任务约10MB
- **CPU开销**：典型工作负载<1%
- **调度延迟**：正常负载下<100ms

## 🔄 最近更新 (v0.2.0-beta)

### 🐛 Bug修复
- 修复集成测试中的竞态条件（`TestCronLongRunningJobs`）
- 纠正基准测试中的无效配置
- 标准化所有测试中的cron表达式格式

### ✨ 增强功能
- **完整集成测试套件**：端到端测试覆盖
- **全面基准测试套件**：所有组件性能测试
- **改进线程安全**：增强同步机制
- **更好的配置管理**：统一配置模式

### 📊 质量改进
- **测试覆盖率**：从60%增加到75.4%
- **测试稳定性**：消除不稳定测试和竞态条件
- **代码质量**：增强错误处理和配置验证

## 🤝 贡献

我们欢迎贡献！请查看我们的[贡献指南](CONTRIBUTING.md)了解详情。

### 开发设置

```bash
# 克隆仓库
git clone https://github.com/callmebg/cron.git
cd cron

# 安装依赖（无需依赖 - 纯Go！）
go mod download

# 运行测试
./scripts/test.sh

# 运行基准测试
./scripts/benchmark.sh

# 检查代码质量
./scripts/lint.sh
```

### 项目指导原则

- **零依赖**：保持纯Go标准库使用
- **高测试覆盖率**：目标>90%测试覆盖率
- **全面文档**：为所有更改更新文档
- **性能关注**：基准测试关键路径

## 📄 许可证

本项目采用MIT许可证 - 查看[LICENSE](LICENSE)文件了解详情。

## 🌟 致谢

- 受Unix cron系统启发
- 使用Go优秀的标准库构建
- 感谢所有贡献者和用户

---

**项目状态**: 🟢 Beta（稳定）- 生产就绪  
**版本**: v0.2.0-beta  
**Go版本**: 1.19+  
**最后更新**: 2025年8月6日  

如有问题、Bug报告或功能请求，请[创建issue](https://github.com/callmebg/cron/issues)。