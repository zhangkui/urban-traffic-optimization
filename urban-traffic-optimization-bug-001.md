# 出题包 BUG-001

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-001 |
| task_type | bugfix |
| bug_category | concurrency |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug1_green
红测分支: bug1_red
基线提交: 77f5ec89860396e77b91b55efb181d8e2ab3b3c1

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-001-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug1_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-001-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-001-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
修复并发交通流量聚合导致路口流量快照损坏的问题。

## 验证命令（verify_cmds）
```bash
go test -race ./internal/traffic -count=1 -run TestRecordBatchConcurrentSensorReports
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：traffic.Service.Record 直接并发写入共享 cache map，缺少与 Recent/LoadCache 一致的互斥保护。生产文件/符号：internal/traffic/service.go 的 Service.Record。调用链：传感器上报 → RecordBatch/Record → Service.cache → Aggregate/Recent。失效原因：多个传感器协程同时执行 map 赋值会触发 data race，极端情况下可能丢失读数或导致运行时并发 map 写崩溃。证据：bug1_red 上 go test -race ./internal/traffic -run TestRecordBatchConcurrentSensorReports 报告 service.go:52 的 concurrent map write。 |

## 成功标准（success_criteria）
目标行为：并发接收同一路口的传感器读数时，所有合法读数都安全写入缓存并可被聚合。边界：不同传感器、相同或不同时间片、重复运行 -race 均不得出现数据竞争；非法读数仍按原校验拒绝。合法场景：8 个采集协程各提交 25 条合法读数，聚合样本数为 200、总流量为 2000。验证标准：回归测试在 -race 下通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug1_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
