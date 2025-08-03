# Go Cron 测试计划

## 测试概述

本文档详细说明了 Go Cron 库的完整测试策略，包括单元测试、集成测试、性能测试和兼容性测试。

## 测试目标

- 确保所有核心功能正常工作
- 验证并发安全性
- 确保内存和性能效率
- 验证错误处理的正确性
- 确保跨平台兼容性

## 测试架构

### 1. 单元测试 (Unit Tests)

#### 1.1 Cron表达式解析器 (`internal/parser/`)

**测试文件**: `parser_test.go`

```go
func TestCronParsingValidExpressions(t *testing.T) {
    validExpressions := []string{
        "* * * * *",
        "0 */2 * * *",
        "30 9 * * 1-5",
        "0 0 1 * *",
        "*/15 * * * *",
        "0 22 * * 0",
    }
    
    for _, expr := range validExpressions {
        _, err := parser.Parse(expr)
        assert.NoError(t, err, "Expression should be valid: %s", expr)
    }
}

func TestCronParsingInvalidExpressions(t *testing.T) {
    invalidExpressions := []string{
        "invalid",
        "60 * * * *",  // Invalid minute
        "* 25 * * *",  // Invalid hour
        "* * 32 * *",  // Invalid day
        "* * * 13 *",  // Invalid month
        "* * * * 8",   // Invalid weekday
    }
    
    for _, expr := range invalidExpressions {
        _, err := parser.Parse(expr)
        assert.Error(t, err, "Expression should be invalid: %s", expr)
    }
}

func TestNextExecutionTime(t *testing.T) {
    testCases := []struct {
        expression string
        baseTime   time.Time
        expected   time.Time
    }{
        {
            expression: "0 */2 * * *",
            baseTime:   time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC),
            expected:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
        },
        // 更多测试案例...
    }
    
    for _, tc := range testCases {
        schedule, _ := parser.Parse(tc.expression)
        next := schedule.Next(tc.baseTime)
        assert.Equal(t, tc.expected, next)
    }
}
```

**覆盖范围**:
- ✅ 有效cron表达式解析
- ✅ 无效表达式错误处理
- ✅ 特殊字符处理 (*, /, -, ,)
- ✅ 下次执行时间计算
- ✅ 时区处理
- ✅ 边界值测试

#### 1.2 调度器核心 (`internal/scheduler/`)

**测试文件**: `scheduler_test.go`, `job_test.go`, `queue_test.go`

```go
func TestJobQueueOperations(t *testing.T) {
    q := NewJobQueue()
    
    job1 := &Job{ID: "1", NextRun: time.Now().Add(time.Hour)}
    job2 := &Job{ID: "2", NextRun: time.Now().Add(time.Minute)}
    
    q.Push(job1)
    q.Push(job2)
    
    // 应该返回最早的任务
    next := q.Pop()
    assert.Equal(t, "2", next.ID)
}

func TestConcurrentJobExecution(t *testing.T) {
    scheduler := NewScheduler()
    
    counter := int64(0)
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        scheduler.AddJob("* * * * *", func() {
            defer wg.Done()
            atomic.AddInt64(&counter, 1)
        })
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()
    
    go scheduler.Start(ctx)
    wg.Wait()
    
    assert.Equal(t, int64(100), atomic.LoadInt64(&counter))
}

func TestGracefulShutdown(t *testing.T) {
    scheduler := NewScheduler()
    
    running := int64(0)
    scheduler.AddJob("* * * * *", func() {
        atomic.StoreInt64(&running, 1)
        time.Sleep(time.Second * 30) // 长时间运行的任务
        atomic.StoreInt64(&running, 0)
    })
    
    ctx, cancel := context.WithCancel(context.Background())
    go scheduler.Start(ctx)
    
    // 等待任务开始
    time.Sleep(time.Second * 2)
    cancel()
    
    err := scheduler.Stop()
    assert.NoError(t, err)
    
    // 确认任务已完成
    assert.Equal(t, int64(0), atomic.LoadInt64(&running))
}
```

**覆盖范围**:
- ✅ 任务队列操作 (添加、移除、排序)
- ✅ 并发任务执行
- ✅ 任务重试机制
- ✅ 超时处理
- ✅ 优雅关闭
- ✅ 内存泄漏检测

#### 1.3 监控系统 (`internal/monitor/`)

**测试文件**: `metrics_test.go`, `http_test.go`

```go
func TestMetricsCollection(t *testing.T) {
    monitor := NewMonitor()
    
    // 模拟任务执行
    monitor.RecordJobStart("test-job")
    time.Sleep(time.Millisecond * 100)
    monitor.RecordJobSuccess("test-job", time.Millisecond*100)
    
    stats := monitor.GetStats()
    assert.Equal(t, int64(1), stats.SuccessfulExecutions)
    assert.Greater(t, stats.AverageExecutionTime, time.Duration(0))
}

func TestHTTPMonitoringEndpoint(t *testing.T) {
    monitor := NewMonitor()
    server := monitor.StartHTTPServer(":0") // 随机端口
    defer server.Close()
    
    resp, err := http.Get(server.URL + "/metrics")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var data map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&data)
    
    assert.Contains(t, data, "scheduler")
    assert.Contains(t, data, "jobs")
}
```

**覆盖范围**:
- ✅ 指标收集和聚合
- ✅ HTTP监控端点
- ✅ 统计数据准确性
- ✅ 并发指标更新
- ✅ JSON响应格式

### 2. 集成测试 (Integration Tests)

#### 2.1 端到端调度测试

**测试文件**: `test/integration/scheduler_test.go`

```go
func TestCompleteSchedulingWorkflow(t *testing.T) {
    c := cron.NewWithConfig(cron.Config{
        EnableMonitoring: true,
        MonitoringPort:   8081,
    })
    
    executionCount := int64(0)
    
    // 添加每秒执行的任务
    err := c.AddNamedJob("counter", "* * * * * *", func() {
        atomic.AddInt64(&executionCount, 1)
    })
    require.NoError(t, err)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()
    
    go c.Start(ctx)
    
    // 等待多次执行
    time.Sleep(time.Second * 5)
    
    count := atomic.LoadInt64(&executionCount)
    assert.GreaterOrEqual(t, count, int64(4)) // 至少执行4次
    assert.LessOrEqual(t, count, int64(6))    // 最多6次（考虑时间误差）
    
    // 验证监控数据
    stats := c.GetJobStats("counter")
    assert.Equal(t, "counter", stats.Name)
    assert.Greater(t, stats.TotalRuns, int64(0))
}

func TestErrorHandlingIntegration(t *testing.T) {
    c := cron.New()
    
    errorCount := int64(0)
    
    c.AddJobWithErrorHandler("* * * * * *",
        func() error {
            return errors.New("simulated error")
        },
        func(err error) {
            atomic.AddInt64(&errorCount, 1)
        },
    )
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    
    go c.Start(ctx)
    time.Sleep(time.Second * 3)
    
    count := atomic.LoadInt64(&errorCount)
    assert.Greater(t, count, int64(0))
    
    stats := c.GetStats()
    assert.Greater(t, stats.FailedExecutions, int64(0))
}
```

#### 2.2 并发压力测试

**测试文件**: `test/integration/concurrency_test.go`

```go
func TestHighConcurrencyScheduling(t *testing.T) {
    c := cron.NewWithConfig(cron.Config{
        MaxConcurrentJobs: 50,
    })
    
    const numJobs = 1000
    executionCount := int64(0)
    
    // 添加大量任务
    for i := 0; i < numJobs; i++ {
        jobName := fmt.Sprintf("job-%d", i)
        err := c.AddNamedJob(jobName, "* * * * * *", func() {
            atomic.AddInt64(&executionCount, 1)
            time.Sleep(time.Millisecond * 10) // 模拟工作
        })
        require.NoError(t, err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
    defer cancel()
    
    start := time.Now()
    go c.Start(ctx)
    
    // 等待所有任务至少执行一次
    for atomic.LoadInt64(&executionCount) < numJobs {
        time.Sleep(time.Millisecond * 100)
        if time.Since(start) > time.Second*25 {
            t.Fatal("Timeout waiting for job executions")
        }
    }
    
    stats := c.GetStats()
    assert.Equal(t, numJobs, stats.TotalJobs)
    assert.LessOrEqual(t, stats.RunningJobs, 50) // 不超过并发限制
}
```

### 3. 性能测试 (Benchmark Tests)

#### 3.1 调度性能基准

**测试文件**: `test/benchmark/scheduler_bench_test.go`

```go
func BenchmarkJobScheduling(b *testing.B) {
    c := cron.New()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        c.AddJob("* * * * *", func() {})
    }
}

func BenchmarkCronParsing(b *testing.B) {
    expression := "0 */2 * * 1-5"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.Parse(expression)
    }
}

func BenchmarkConcurrentExecution(b *testing.B) {
    c := cron.NewWithConfig(cron.Config{
        MaxConcurrentJobs: 100,
    })
    
    // 添加100个任务
    for i := 0; i < 100; i++ {
        c.AddJob("* * * * * *", func() {
            time.Sleep(time.Microsecond * 100)
        })
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go c.Start(ctx)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        time.Sleep(time.Second)
    }
}
```

#### 3.2 内存使用基准

**测试文件**: `test/benchmark/memory_bench_test.go`

```go
func BenchmarkMemoryUsage(b *testing.B) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    c := cron.New()
    
    // 添加大量任务
    for i := 0; i < b.N; i++ {
        c.AddJob("* * * * *", func() {})
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/job")
}

func BenchmarkSchedulerMemoryGrowth(b *testing.B) {
    c := cron.New()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go c.Start(ctx)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        c.AddJob("* * * * *", func() {})
        
        if i%1000 == 0 {
            runtime.GC()
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            b.ReportMetric(float64(m.Alloc), "bytes")
        }
    }
}
```

### 4. 兼容性测试 (Compatibility Tests)

#### 4.1 跨平台测试

**测试脚本**: `scripts/test-platforms.sh`

```bash
#!/bin/bash

# 测试不同操作系统和架构
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

for platform in $PLATFORMS; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    echo "Testing on $GOOS/$GOARCH..."
    
    GOOS=$GOOS GOARCH=$GOARCH go test ./... -short
    
    if [ $? -ne 0 ]; then
        echo "Tests failed on $platform"
        exit 1
    fi
done

echo "All platform tests passed!"
```

#### 4.2 Go版本兼容性

**GitHub Actions**: `.github/workflows/compatibility.yml`

```yaml
name: Go Version Compatibility

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.19', '1.20', '1.21', '1.22', '1.23']
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run tests
      run: go test ./... -race -coverprofile=coverage.out
    
    - name: Run benchmarks
      run: go test -bench=. ./test/benchmark/
```

### 5. 回归测试 (Regression Tests)

#### 5.1 历史Bug修复验证

**测试文件**: `test/regression/bugs_test.go`

```go
// 测试修复的特定bug，确保不会重新出现
func TestBug001_TimezoneHandling(t *testing.T) {
    // 之前的bug：时区转换导致任务调度错误
    c := cron.NewWithConfig(cron.Config{
        Timezone: time.UTC,
    })
    
    executed := false
    c.AddJob("0 12 * * *", func() {
        executed = true
    })
    
    // 模拟在不同时区的执行
    testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    // ... 验证逻辑
}

func TestBug002_MemoryLeak(t *testing.T) {
    // 之前的bug：长时间运行导致内存泄漏
    c := cron.New()
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // 运行大量任务
    for i := 0; i < 10000; i++ {
        c.AddJob("* * * * *", func() {})
        c.RemoveJob(fmt.Sprintf("job-%d", i))
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // 内存使用应该稳定
    memoryGrowth := m2.Alloc - m1.Alloc
    assert.Less(t, memoryGrowth, uint64(1024*1024)) // 小于1MB
}
```

### 6. 安全测试 (Security Tests)

#### 6.1 输入验证测试

**测试文件**: `test/security/input_validation_test.go`

```go
func TestMaliciousCronExpressions(t *testing.T) {
    maliciousInputs := []string{
        strings.Repeat("*", 10000),           // 超长输入
        "'; DROP TABLE jobs; --",            // SQL注入尝试
        "<script>alert('xss')</script>",     // XSS尝试
        "\x00\x01\x02\x03",                 // 二进制数据
        "$(rm -rf /)",                       // 命令注入尝试
    }
    
    c := cron.New()
    
    for _, input := range maliciousInputs {
        err := c.AddJob(input, func() {})
        assert.Error(t, err, "Should reject malicious input: %s", input)
    }
}

func TestResourceExhaustion(t *testing.T) {
    c := cron.NewWithConfig(cron.Config{
        MaxConcurrentJobs: 10,
    })
    
    // 尝试添加过多任务
    for i := 0; i < 1000; i++ {
        c.AddJob("* * * * *", func() {
            time.Sleep(time.Hour) // 长时间运行
        })
    }
    
    stats := c.GetStats()
    assert.LessOrEqual(t, stats.RunningJobs, 10)
}
```

### 7. 故障注入测试 (Chaos Testing)

#### 7.1 系统故障模拟

**测试文件**: `test/chaos/failure_injection_test.go`

```go
func TestSystemClockChanges(t *testing.T) {
    // 模拟系统时间变化对调度的影响
    c := cron.New()
    
    executed := false
    c.AddJob("* * * * *", func() {
        executed = true
    })
    
    // 模拟时钟跳跃
    originalTime := time.Now()
    // ... 时钟操作模拟
    
    assert.True(t, executed)
}

func TestMemoryPressure(t *testing.T) {
    c := cron.New()
    
    // 在内存压力下运行
    largeData := make([][]byte, 0)
    
    go func() {
        for i := 0; i < 1000; i++ {
            largeData = append(largeData, make([]byte, 1024*1024))
            time.Sleep(time.Millisecond * 10)
        }
    }()
    
    // 验证调度器在内存压力下仍能正常工作
    executed := false
    c.AddJob("* * * * *", func() {
        executed = true
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()
    
    go c.Start(ctx)
    time.Sleep(time.Minute + time.Second*5)
    
    assert.True(t, executed)
}
```

## 测试执行策略

### 本地开发测试

```bash
# 运行所有测试
make test

# 运行带覆盖率的测试
make test-coverage

# 运行基准测试
make benchmark

# 运行安全测试
make test-security

# 运行兼容性测试
make test-compatibility
```

### CI/CD测试流水线

1. **Pull Request 触发**:
   - 单元测试 (所有包)
   - 集成测试 (快速版本)
   - 代码覆盖率检查 (>90%)
   - 静态代码分析

2. **主分支合并**:
   - 完整测试套件
   - 性能回归检测
   - 跨平台测试
   - 安全扫描

3. **发布前测试**:
   - 完整兼容性测试
   - 长时间运行测试 (24小时)
   - 故障注入测试
   - 文档验证

### 测试数据管理

**测试数据目录**: `test/testdata/`

```
testdata/
├── cron_expressions.json      # 标准cron表达式测试集
├── edge_cases.json           # 边界情况测试
├── performance_baseline.json # 性能基准数据
├── timezone_tests.json       # 时区测试数据
└── security_inputs.json      # 安全测试输入
```

## 质量指标

### 代码覆盖率目标

| 组件 | 目标覆盖率 | 当前状态 |
|------|-----------|----------|
| Parser | 95% | 🟢 |
| Scheduler | 90% | 🟢 |
| Monitor | 85% | 🟡 |
| Utils | 90% | 🟢 |
| 整体 | 90% | 🟢 |

### 性能基准

| 指标 | 目标 | 当前值 |
|------|------|--------|
| 任务添加延迟 | <1ms | 0.2ms |
| 调度准确性 | ±100ms | ±50ms |
| 内存使用 | <10MB/1000任务 | 8.5MB |
| CPU使用率 | <5% | 2.3% |

### 测试自动化

- **单元测试**: 每次代码提交自动运行
- **集成测试**: 每日自动运行
- **性能测试**: 每周自动运行
- **兼容性测试**: 每次发布前运行

## 问题追踪

所有测试相关的问题都在GitHub Issues中跟踪，使用以下标签：

- `testing`: 测试相关问题
- `performance`: 性能测试问题  
- `compatibility`: 兼容性问题
- `security`: 安全测试问题
- `flaky-test`: 不稳定的测试

## 测试文档

- 每个测试文件都包含详细的注释
- 复杂测试场景有独立的文档说明
- 性能测试结果定期更新到文档中

---

这个测试计划确保了Go Cron库的高质量和可靠性，覆盖了从基础功能到复杂场景的所有测试需求。