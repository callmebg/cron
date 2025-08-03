# Go Cron - 原生定时任务调度器

一个轻量级、高效的Go应用程序定时任务调度器，完全基于Go标准库构建。该库提供强大的任务调度功能和全面的监控能力，同时保持零外部依赖。

[English](README.md) | 中文

## 特性

- **纯Go标准库**：无外部依赖，充分利用Go的原生能力
- **标准Cron语法**：完全支持传统cron表达式
- **并发执行**：使用goroutines和channels高效执行任务
- **内置监控**：全面的指标和执行跟踪
- **优雅关闭**：基于context的清理终止
- **线程安全**：支持多个goroutines并发使用
- **灵活日志**：与标准`log`包集成的可配置日志
- **内存高效**：最小内存占用和高效调度算法

## 项目结构

项目遵循Go约定，组织清晰，易于维护和贡献：

```
cron/
├── cmd/
│   └── examples/
│       ├── basic/
│       │   └── main.go              # 基础使用示例
│       ├── advanced/
│       │   └── main.go              # 高级配置示例
│       ├── monitoring/
│       │   └── main.go              # 监控和指标示例
│       └── testing/
│           └── main.go              # 测试模式示例
├── internal/
│   ├── parser/
│   │   ├── parser.go                # Cron表达式解析器
│   │   ├── parser_test.go           # 解析器单元测试
│   │   └── validator.go             # 表达式验证
│   ├── scheduler/
│   │   ├── scheduler.go             # 核心调度逻辑
│   │   ├── scheduler_test.go        # 调度器测试
│   │   ├── job.go                   # 任务定义和管理
│   │   ├── job_test.go              # 任务测试
│   │   └── queue.go                 # 优先队列实现
│   ├── monitor/
│   │   ├── metrics.go               # 指标收集
│   │   ├── metrics_test.go          # 指标测试
│   │   ├── http.go                  # HTTP监控端点
│   │   └── stats.go                 # 统计信息聚合
│   └── utils/
│       ├── time.go                  # 时间工具
│       ├── time_test.go             # 时间工具测试
│       └── sync.go                  # 同步辅助工具
├── pkg/
│   └── cron/
│       ├── cron.go                  # 公共API和主入口点
│       ├── cron_test.go             # 集成测试
│       ├── config.go                # 配置类型和默认值
│       ├── config_test.go           # 配置测试
│       ├── errors.go                # 错误定义
│       └── types.go                 # 公共类型定义
├── test/
│   ├── integration/
│   │   ├── scheduler_test.go        # 集成测试
│   │   └── monitoring_test.go       # 监控集成测试
│   ├── benchmark/
│   │   ├── scheduler_bench_test.go  # 性能基准测试
│   │   └── memory_bench_test.go     # 内存使用基准测试
│   └── testdata/
│       ├── cron_expressions.txt     # 测试cron表达式
│       └── expected_schedules.json  # 预期调度结果
├── docs/
│   ├── architecture.md              # 架构概述
│   ├── performance.md               # 性能特征
│   └── contributing.md              # 详细贡献指南
├── scripts/
│   ├── test.sh                      # 综合测试脚本
│   ├── benchmark.sh                 # 基准测试脚本
│   └── lint.sh                      # 代码质量检查
├── .github/
│   └── workflows/
│       ├── ci.yml                   # 持续集成
│       ├── release.yml              # 发布自动化
│       └── security.yml             # 安全扫描
├── go.mod                           # Go模块定义
├── go.sum                           # 依赖校验和
├── LICENSE                          # MIT许可证
├── README.md                        # 英文文档
├── README_zh.md                     # 中文文档（本文件）
├── CHANGELOG.md                     # 版本历史
└── Makefile                         # 构建自动化

```

### 目录说明

#### `/cmd/examples/`
包含展示不同用例的实际示例：
- **basic/**：简单的cron任务设置和执行
- **advanced/**：复杂配置和错误处理
- **monitoring/**：HTTP监控和指标收集
- **testing/**：测试模式和最佳实践

#### `/internal/`
不对外暴露的私有实现包：
- **parser/**：Cron表达式解析和验证逻辑
- **scheduler/**：核心调度引擎和任务队列管理
- **monitor/**：指标收集和HTTP监控端点
- **utils/**：时间处理和同步的共享工具

#### `/pkg/cron/`
用户导入的公共API包：
- **cron.go**：主调度器接口和公共方法
- **config.go**：配置结构和默认值
- **types.go**：公共类型定义和接口
- **errors.go**：公共错误类型和错误处理

#### `/test/`
全面的测试套件：
- **integration/**：端到端集成测试
- **benchmark/**：性能和内存基准测试
- **testdata/**：测试固件和预期结果

#### `/docs/`
除README外的额外文档：
- **architecture.md**：内部架构和设计决策
- **performance.md**：性能特征和优化说明
- **contributing.md**：详细的开发和贡献指南

#### `/scripts/`
开发和CI自动化脚本：
- **test.sh**：运行所有测试并生成覆盖率报告
- **benchmark.sh**：执行性能基准测试
- **lint.sh**：代码质量和格式检查

#### `/.github/workflows/`
CI/CD的GitHub Actions：
- **ci.yml**：跨Go版本和平台的自动化测试
- **release.yml**：自动化发布和标记
- **security.yml**：安全漏洞扫描

### 包组织原则

1. **清晰分离**：`/pkg/cron/`中的公共API，`/internal/`中的实现
2. **单一职责**：每个包都有专一的、明确定义的目的
3. **可测试性**：所有包都设计为易于单元和集成测试
4. **文档完善**：所有公共API都有全面的示例和文档
5. **仅标准库**：零外部依赖，充分利用Go的内置能力

### 导入路径结构

```go
// 公共API - 用户导入的内容
import "github.com/callmebg/cron/pkg/cron"

// 内部包 - 用户无法访问
// internal/parser
// internal/scheduler  
// internal/monitor
// internal/utils
```

这种结构确保了公共API和实现细节的清晰分离，同时提供全面的测试、文档和示例。

## 安装

```bash
go get github.com/callmebg/cron
```

## 快速开始

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
    // 创建新的cron调度器
    c := cron.New()

    // 添加每分钟运行的任务
    err := c.AddJob("* * * * *", func() {
        fmt.Println("任务执行时间:", time.Now())
    })
    if err != nil {
        log.Fatal(err)
    }

    // 启动调度器
    ctx := context.Background()
    c.Start(ctx)

    // 保持程序运行
    select {}
}
```

## Cron语法

该库支持标准的5字段cron语法：

```
┌─────────────── 分钟 (0 - 59)
│ ┌───────────── 小时 (0 - 23)
│ │ ┌─────────── 日期 (1 - 31)
│ │ │ ┌───────── 月份 (1 - 12)
│ │ │ │ ┌─────── 星期 (0 - 6) (星期日到星期六)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

### 特殊字符

- `*` - 任意值
- `,` - 值列表分隔符
- `-` - 值范围
- `/` - 步长值

### 示例

| 表达式 | 描述 |
|------------|-------------|
| `0 */2 * * *` | 每2小时 |
| `30 9 * * 1-5` | 工作日上午9:30 |
| `0 0 1 * *` | 每月第一天 |
| `*/15 * * * *` | 每15分钟 |
| `0 22 * * 0` | 每周日晚上10点 |

## 使用示例

### 基础任务调度

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    c := cron.New()

    // 简单周期性任务
    c.AddJob("*/5 * * * *", func() {
        fmt.Println("这个任务每5分钟运行一次")
    })

    // 命名任务便于监控
    c.AddNamedJob("backup", "0 2 * * *", func() {
        fmt.Println("运行夜间备份")
        // 你的备份逻辑
    })

    // 带错误处理的任务
    c.AddJobWithErrorHandler("0 */6 * * *", 
        func() error {
            // 你的任务逻辑
            return doSomeWork()
        },
        func(err error) {
            log.Printf("任务失败: %v", err)
        },
    )

    c.Start(context.Background())
    select {}
}

func doSomeWork() error {
    // 模拟可能失败的工作
    return nil
}
```

### 高级配置

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // 自定义配置
    config := cron.Config{
        Logger:           log.New(os.Stdout, "CRON: ", log.LstdFlags),
        Timezone:         time.UTC,
        MaxConcurrentJobs: 10,
        EnableMonitoring: true,
        MonitoringPort:   8080,
    }

    c := cron.NewWithConfig(config)

    // 添加不同配置的任务
    jobConfig := cron.JobConfig{
        MaxRetries:    3,
        RetryInterval: time.Minute * 5,
        Timeout:       time.Hour,
    }

    c.AddJobWithConfig("0 */1 * * *", jobConfig, func() {
        // 这个任务失败时将重试最多3次
        // 并且会在1小时后超时
    })

    c.Start(context.Background())
    select {}
}
```

## 监控能力

### 内置指标

该库开箱即用地提供全面监控：

```go
// 获取调度器统计信息
stats := c.GetStats()
fmt.Printf("总任务数: %d\n", stats.TotalJobs)
fmt.Printf("运行中任务: %d\n", stats.RunningJobs)
fmt.Printf("成功执行次数: %d\n", stats.SuccessfulExecutions)
fmt.Printf("失败执行次数: %d\n", stats.FailedExecutions)
fmt.Printf("平均执行时间: %v\n", stats.AverageExecutionTime)
```

### 任务级监控

```go
// 获取特定任务的统计信息
jobStats := c.GetJobStats("backup")
fmt.Printf("最后执行: %v\n", jobStats.LastExecution)
fmt.Printf("下次执行: %v\n", jobStats.NextExecution)
fmt.Printf("成功率: %.2f%%\n", jobStats.SuccessRate)
fmt.Printf("总运行次数: %d\n", jobStats.TotalRuns)
```

### HTTP监控端点

启用监控时，调度器会暴露HTTP端点：

```go
c := cron.NewWithConfig(cron.Config{
    EnableMonitoring: true,
    MonitoringPort:   8080,
})
```

在 `http://localhost:8080/metrics` 访问监控数据：

```json
{
  "scheduler": {
    "uptime": "2h45m30s",
    "total_jobs": 5,
    "running_jobs": 2,
    "successful_executions": 127,
    "failed_executions": 3
  },
  "jobs": [
    {
      "name": "backup",
      "schedule": "0 2 * * *",
      "last_execution": "2024-01-15T02:00:00Z",
      "next_execution": "2024-01-16T02:00:00Z",
      "status": "completed",
      "execution_time": "45s",
      "success_rate": 98.5
    }
  ]
}
```

## API参考

### 核心类型

#### 调度器

```go
type Scheduler struct {
    // 内部字段
}

// 使用默认配置创建新调度器
func New() *Scheduler

// 使用自定义配置创建新调度器
func NewWithConfig(config Config) *Scheduler

// 启动调度器
func (s *Scheduler) Start(ctx context.Context)

// 优雅停止调度器
func (s *Scheduler) Stop() error

// 使用cron表达式添加任务
func (s *Scheduler) AddJob(schedule string, job func()) error

// 添加命名任务便于监控
func (s *Scheduler) AddNamedJob(name, schedule string, job func()) error

// 添加带错误处理的任务
func (s *Scheduler) AddJobWithErrorHandler(schedule string, job func() error, errorHandler func(error)) error

// 添加带自定义配置的任务
func (s *Scheduler) AddJobWithConfig(schedule string, config JobConfig, job func()) error

// 按名称移除任务
func (s *Scheduler) RemoveJob(name string) error

// 获取调度器统计信息
func (s *Scheduler) GetStats() Stats

// 获取任务特定统计信息
func (s *Scheduler) GetJobStats(name string) JobStats
```

#### 配置

```go
type Config struct {
    Logger           *log.Logger
    Timezone         *time.Location
    MaxConcurrentJobs int
    EnableMonitoring bool
    MonitoringPort   int
}

type JobConfig struct {
    MaxRetries    int
    RetryInterval time.Duration
    Timeout       time.Duration
}
```

#### 统计信息

```go
type Stats struct {
    TotalJobs              int
    RunningJobs            int
    SuccessfulExecutions   int64
    FailedExecutions       int64
    AverageExecutionTime   time.Duration
    Uptime                 time.Duration
}

type JobStats struct {
    Name            string
    Schedule        string
    LastExecution   time.Time
    NextExecution   time.Time
    Status          string
    ExecutionTime   time.Duration
    SuccessRate     float64
    TotalRuns       int64
}
```

## 配置选项

### 调度器配置

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `Logger` | `*log.Logger` | `log.Default()` | 调度器事件的日志实例 |
| `Timezone` | `*time.Location` | `time.Local` | 调度评估的时区 |
| `MaxConcurrentJobs` | `int` | `100` | 最大并发任务执行数 |
| `EnableMonitoring` | `bool` | `false` | 启用HTTP监控端点 |
| `MonitoringPort` | `int` | `8080` | 监控HTTP服务器端口 |

### 任务配置

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `MaxRetries` | `int` | `0` | 最大重试次数 |
| `RetryInterval` | `time.Duration` | `time.Minute` | 重试间隔 |
| `Timeout` | `time.Duration` | `0` (无超时) | 每个任务的最大执行时间 |

## 最佳实践

### 1. 使用Context进行取消

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c.Start(ctx)

// 优雅关闭
go func() {
    <-sigterm
    cancel()
    c.Stop()
}()
```

### 2. 正确处理错误

```go
c.AddJobWithErrorHandler("*/5 * * * *",
    func() error {
        return performTask()
    },
    func(err error) {
        log.Printf("任务失败: %v", err)
        // 发送警报，记录到外部系统等
    },
)
```

### 3. 使用命名任务进行监控

```go
// 好：命名任务便于识别
c.AddNamedJob("database-cleanup", "0 3 * * *", cleanupDatabase)

// 更好：带配置
c.AddJobWithConfig("user-sync", "*/30 * * * *", 
    cron.JobConfig{
        MaxRetries: 3,
        Timeout: time.Minute * 10,
    },
    syncUsers,
)
```

### 4. 资源管理

```go
func resourceIntensiveJob() {
    // 限制资源使用
    semaphore := make(chan struct{}, 5) // 最多5个并发操作
    
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(item string) {
            defer wg.Done()
            semaphore <- struct{}{}        // 获取
            defer func() { <-semaphore }() // 释放
            
            processItem(item)
        }(item)
    }
    wg.Wait()
}
```

### 5. 测试Cron任务

```go
func TestJobExecution(t *testing.T) {
    c := cron.New()
    
    executed := false
    err := c.AddJob("* * * * *", func() {
        executed = true
    })
    require.NoError(t, err)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()
    
    go c.Start(ctx)
    
    // 等待执行
    time.Sleep(time.Minute + time.Second*5)
    assert.True(t, executed)
}
```

## 性能考虑

### 内存使用

- 调度器使用基于堆的优先队列进行高效任务调度
- 内存使用与调度任务数量呈线性增长
- 每个任务维持最少的元数据用于监控

### CPU使用

- 调度开销最小，使用基于时间的触发器
- 任务执行在独立的goroutines中进行，防止阻塞
- 调度器在评估间隔之间休眠以最小化CPU使用

### 并发性

- 支持多个goroutines并发访问安全
- 使用sync.RWMutex进行线程安全操作
- 可配置的最大并发任务执行限制

## 错误处理

该库提供多级错误处理：

1. **调度解析错误**：添加任务时立即返回
2. **任务执行错误**：通过错误处理器处理或记录
3. **系统错误**：记录并通过监控报告

综合错误处理示例：

```go
c := cron.NewWithConfig(cron.Config{
    Logger: log.New(os.Stderr, "CRON-ERROR: ", log.LstdFlags),
})

// 处理调度解析错误
if err := c.AddJob("invalid cron", myJob); err != nil {
    log.Fatalf("无效的cron表达式: %v", err)
}

// 处理任务执行错误
c.AddJobWithErrorHandler("0 */1 * * *",
    func() error {
        return riskyOperation()
    },
    func(err error) {
        // 自定义错误处理
        notifyAdmin(err)
        logToExternalSystem(err)
    },
)
```

## 贡献

我们欢迎贡献！请遵循以下准则：

### 开发设置

1. 克隆仓库：
   ```bash
   git clone https://github.com/callmebg/cron.git
   cd cron
   ```

2. 运行测试：
   ```bash
   go test ./...
   ```

3. 运行基准测试：
   ```bash
   go test -bench=. ./...
   ```

### 代码标准

- 遵循Go格式化约定 (`gofmt`)
- 为新功能编写全面的测试
- 更新API变更的文档
- 仅使用Go标准库依赖
- 保持向后兼容性

### 提交变更

1. Fork仓库
2. 创建功能分支：`git checkout -b feature-name`
3. 提交你的变更：`git commit -am 'Add feature'`
4. 推送到分支：`git push origin feature-name`
5. 提交拉取请求

### 测试

所有贡献必须包含测试：

```go
func TestNewFeature(t *testing.T) {
    c := cron.New()
    
    // 测试你的功能
    err := c.YourNewMethod()
    assert.NoError(t, err)
}

func BenchmarkNewFeature(b *testing.B) {
    c := cron.New()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        c.YourNewMethod()
    }
}
```

## 许可证

本项目在MIT许可证下授权 - 详见 [LICENSE](LICENSE) 文件。

## 测试

本项目遵循全面的测试实践，确保可靠性和性能。详细的测试策略请查看我们的[测试计划](TEST_PLAN.md)。

### 测试覆盖率

- **单元测试**: 所有包90%+覆盖率
- **集成测试**: 端到端调度工作流程
- **基准测试**: 性能和内存使用验证  
- **兼容性测试**: 跨平台和Go版本测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -coverprofile=coverage.out ./...

# 运行基准测试
go test -bench=. ./test/benchmark/

# 运行集成测试
go test ./test/integration/
```

## 与其他库的对比

我们与其他流行的Go cron库进行了详细对比。完整分析请查看[COMPARISON.md](COMPARISON.md)。

### 快速对比

| 特性 | callmebg/cron | robfig/cron | go-co-op/gocron |
|------|--------------|-------------|-----------------|
| 零依赖 | ✅ | ✅ | ❌ |
| 内置监控 | ✅ | ❌ | ⚠️ |
| 错误处理 | ✅ 高级 | ⚠️ 基础 | ✅ 良好 |
| 并发控制 | ✅ | ❌ | ⚠️ |
| 生产就绪 | ✅ | ✅ | ✅ |

### 为什么选择 callmebg/cron？

- **零依赖**: 仅使用Go标准库构建
- **生产就绪**: 全面的监控和错误处理
- **高性能**: 为低内存使用和高吞吐量优化
- **开发友好**: 清晰的API和详细的文档
- **企业特性**: 重试机制、超时和优雅关闭

## 路线图

- [ ] 支持秒级精度（6字段cron语法）
- [ ] 分布式调度能力
- [ ] 任务管理Web UI
- [ ] Prometheus指标导出
- [ ] 跨重启任务持久化
- [ ] 动态任务修改
- [ ] 任务依赖管理

## 支持

- 为bug报告或功能请求创建issue
- 在创建新issue前检查现有issue
- 为bug提供最小复现案例
- 在bug报告中包含Go版本和操作系统信息

---

使用Go标准库构建 - 无需外部依赖。