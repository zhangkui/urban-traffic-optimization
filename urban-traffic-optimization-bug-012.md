# 出题包 BUG-012

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-012 |
| task_type | diagnosis |
| bug_category | nil |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug12_green
红测分支: bug12_red
基线提交: b9518ac4e5764d63efce09799a0d1015d97e1a87

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-012-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug12_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-012-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-012-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
路口没有交通样本时，指标计算必须返回安全的空结果。

## 验证命令（verify_cmds）
```bash
go test ./internal/analysis -count=1 -run TestEmptySampleSummaryReturnsZeroMetric
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：空交通样本时 EmptySampleSummary 将聚合指针保持为 nil，却直接读取其 Volume、AverageSpeed 和 AverageDelay 字段。生产文件/符号：internal/analysis/service.go 的 Service.EmptySampleSummary。调用链：路口指标查询 → EmptySampleSummary → traffic.Aggregate → nil 聚合指针字段访问。失效原因：新路口、设备离线或时间范围无数据时没有可用样本，指标接口应返回零值而不是解引用 nil。证据：bug12_red 回归测试对无流量样本的路口调用方法，报告 invalid memory address or nil pointer dereference。 |

## 成功标准（success_criteria）
目标行为：没有交通样本的路口指标返回安全零值。边界：无样本、单样本、多样本和不同时间窗口都不得崩溃；拥堵指数和延误应保持可解释的零值。合法场景：空时间范围查询返回 Volume、Speed、Delay 均为 0。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug12_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
