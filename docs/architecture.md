# Go Cron 架构设计文档

## 架构概览

Go Cron 采用模块化的架构设计，将功能分解为独立的包，每个包负责特定的职责。整体架构遵循清晰的分层原则和依赖倒置原则。

## 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    用户应用层                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   基础示例   │  │   高级配置   │  │   监控集成   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                 公共API层 (pkg/cron/)                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  cron.go │ config.go │ types.go │ errors.go          │   │
│  │          Scheduler 主接口和配置管理                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   内部实现层 (internal/)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   parser/   │  │ scheduler/  │  │  monitor/   │        │
│  │  表达式解析  │  │ 任务调度核心 │  │  监控指标   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│  ┌─────────────┐  ┌─────────────┐                         │
│  │   types/    │  │   utils/    │                         │
│  │  内部类型   │  │  工具函数   │                         │
│  └─────────────┘  └─────────────┘                         │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go 标准库                              │
│  time, context, sync, log, net/http, encoding/json...      │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件详解

### 1. 调度器核心 (pkg/cron/cron.go)

**职责**: 主调度器，协调所有组件工作

**核心功能**:
- 调度器生命周期管理 (Start/Stop)
- 任务的添加、删除和管理
- 并发控制和资源管理
- 统计信息收集

**关键数据结构**:
```go
type Scheduler struct {
    config    Config              // 调度器配置
    queue     *scheduler.JobQueue // 任务优先队列
    running   bool               // 运行状态
    stopCh    chan struct{}      // 停止信号通道
    mu        sync.RWMutex       // 读写锁
    jobIDGen  int64              // 任务ID生成器
    jobs      map[string]*scheduler.Job // 任务映射
    stats     Stats              // 统计信息
    startTime time.Time          // 启动时间
}
```

**调度循环**:
```go
func (s *Scheduler) run(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-s.stopCh:
            return
        case now := <-ticker.C:
            s.processReadyJobs(ctx, now)
        }
    }
}
```

### 2. Cron表达式解析器 (internal/parser/)

**职责**: 解析和验证cron表达式，计算下次执行时间

**核心算法**:
- 5字段cron表达式解析 (分钟、小时、日、月、周)
- 支持通配符、范围、步长、列表语法
- 时区感知的时间计算

**解析流程**:
```
输入: "0 */2 * * 1-5"
  ↓
字段分割: ["0", "*/2", "*", "*", "1-5"]
  ↓
字段解析: [分钟:0] [小时:0,2,4,6,8,10,12,14,16,18,20,22] [日:*] [月:*] [周:1,2,3,4,5]
  ↓
调度对象: Schedule{...}
```

**时间计算算法**:
```go
func (s *Schedule) Next(t time.Time) time.Time {
    // 从给定时间开始，逐分钟检查是否匹配所有字段
    // 使用高效的位运算和查找表优化
    for {
        if s.matches(t) {
            return t
        }
        t = t.Add(time.Minute)
    }
}
```

### 3. 任务队列 (internal/scheduler/queue.go)

**职责**: 高效管理待执行任务的优先队列

**数据结构**: 最小堆 (Min-Heap)
```go
type JobQueue struct {
    jobs []*Job        // 堆数组
    mu   sync.RWMutex  // 读写锁
}
```

**核心操作**:
- `Add(job)`: O(log n) 时间复杂度添加任务
- `PopReady(now)`: O(k log n) 获取所有准备执行的任务
- `Remove(jobID)`: O(n) 删除指定任务

**堆维护算法**:
```go
func (q *JobQueue) heapifyUp(index int) {
    parent := (index - 1) / 2
    if parent >= 0 && q.jobs[parent].NextRun.After(q.jobs[index].NextRun) {
        q.jobs[parent], q.jobs[index] = q.jobs[index], q.jobs[parent]
        q.heapifyUp(parent)
    }
}
```

### 4. 任务模型 (internal/scheduler/job.go)

**职责**: 封装单个任务的完整生命周期

**任务状态机**:
```
Created → Scheduled → Running → Completed/Failed → Scheduled
    ↓         ↓         ↓            ↓
  NextRun   Queue    Execute    UpdateStats
```

**核心功能**:
- 任务执行和状态管理
- 重试机制和超时控制
- 统计信息收集
- 错误处理

**执行流程**:
```go
func (j *Job) Execute(ctx context.Context) error {
    // 1. 超时控制
    if j.Config.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, j.Config.Timeout)
        defer cancel()
    }
    
    // 2. 执行任务
    j.IsRunning = true
    start := time.Now()
    err := j.executeWithRetry(ctx)
    j.LastDuration = time.Since(start)
    j.IsRunning = false
    
    // 3. 更新统计
    j.updateStats(err)
    
    return err
}
```

### 5. 监控系统 (internal/monitor/)

**职责**: 实时收集和暴露调度器的监控指标

**监控架构**:
```
Scheduler Events → Metrics Collector → Aggregator → HTTP Endpoint
                                    ↘               ↗
                                     → Memory Store →
```

**指标类型**:
- **计数器**: 执行次数、失败次数
- **直方图**: 执行时间分布
- **仪表盘**: 当前运行任务数、队列长度
- **摘要**: 平均执行时间、成功率

**HTTP监控端点**:
```
GET /metrics
{
  "scheduler": {
    "uptime": "2h45m30s",
    "total_jobs": 15,
    "running_jobs": 3,
    "successful_executions": 127,
    "failed_executions": 8
  },
  "jobs": [...]
}
```

## 设计原则

### 1. 单一职责原则 (SRP)
每个包都有明确的单一职责：
- `parser`: 专注cron表达式解析
- `scheduler`: 专注任务调度逻辑
- `monitor`: 专注监控指标收集
- `utils`: 提供通用工具函数

### 2. 开闭原则 (OCP)
- 通过接口和配置实现扩展性
- 新增功能不需要修改现有核心代码
- 支持插件式的错误处理器

### 3. 依赖倒置原则 (DIP)
- 高层模块不依赖低层模块细节
- 都依赖于抽象接口
- 便于测试和模块替换

### 4. 接口隔离原则 (ISP)
- 提供最小必要的公共接口
- 内部实现细节完全隐藏
- 用户只需了解公共API

## 并发模型

### 1. 调度器主循环
```
Main Goroutine: 调度循环 (1个)
    ├── 每秒检查待执行任务
    ├── 处理优先队列
    └── 分发任务到工作goroutine
```

### 2. 任务执行
```
Worker Goroutines: 任务执行 (可配置数量)
    ├── 并发限制: MaxConcurrentJobs
    ├── 独立执行: 不阻塞调度循环
    └── 完成后: 更新统计，重新排队
```

### 3. 监控系统
```
Monitor Goroutine: HTTP服务器 (1个)
    ├── 非阻塞监控端点
    ├── 实时统计聚合
    └── JSON API响应
```

### 4. 同步机制
- **sync.RWMutex**: 保护共享数据结构
- **channel**: goroutine间通信
- **context**: 优雅关闭和超时控制
- **atomic**: 高频统计计数器

## 内存管理

### 1. 对象池化
```go
// 任务对象复用，减少GC压力
var jobPool = sync.Pool{
    New: func() interface{} {
        return &Job{}
    },
}
```

### 2. 内存优化策略
- **紧凑数据结构**: 最小化内存占用
- **及时释放**: 完成的任务立即清理
- **池化复用**: 高频对象使用对象池
- **分代GC友好**: 减少长期持有的短生命周期对象

### 3. 内存使用模式
```
Base Memory: ~2MB (调度器基础开销)
Per Job: ~800 bytes (包含统计信息)
Queue Overhead: O(n) 线性增长
Monitor Buffer: ~1MB (监控数据缓存)
```

## 错误处理策略

### 1. 分层错误处理
```
用户层: JobFuncWithError → ErrorHandler
  ↓
调度层: 任务执行错误 → 重试机制
  ↓
系统层: 调度器错误 → 日志记录
  ↓
底层: panic恢复 → 优雅降级
```

### 2. 错误分类
- **预期错误**: 业务逻辑错误，正常处理
- **重试错误**: 临时性错误，自动重试
- **严重错误**: 系统级错误，记录并告警
- **致命错误**: 不可恢复，优雅关闭

### 3. 容错机制
- **重试**: 指数退避重试策略
- **熔断**: 连续失败后暂停任务
- **隔离**: 单个任务失败不影响其他任务
- **恢复**: 系统重启后自动恢复调度

## 扩展点

### 1. 自定义调度算法
```go
type ScheduleCalculator interface {
    Next(current time.Time) time.Time
}
```

### 2. 插件式监控
```go
type MetricsCollector interface {
    RecordExecution(jobName string, duration time.Duration, err error)
    GetStats() map[string]interface{}
}
```

### 3. 可扩展的存储后端
```go
type JobStore interface {
    Save(job *Job) error
    Load(jobID string) (*Job, error)
    Delete(jobID string) error
}
```

## 性能特征

### 1. 时间复杂度
- 添加任务: O(log n)
- 获取待执行任务: O(k log n), k为待执行任务数
- 删除任务: O(n)
- cron表达式解析: O(1)

### 2. 空间复杂度
- 任务存储: O(n)
- 优先队列: O(n)
- 监控缓存: O(1)

### 3. 性能优化
- **预计算**: cron表达式预解析
- **批处理**: 批量获取待执行任务
- **缓存**: 监控数据智能缓存
- **无锁操作**: 高频计数器使用atomic

这个架构设计确保了Go Cron库的高性能、可扩展性和可维护性，适合各种规模的生产环境使用。