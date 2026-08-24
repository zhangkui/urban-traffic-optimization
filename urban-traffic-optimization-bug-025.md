# 出题包 BUG-025

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-025 |
| task_type | bugfix |
| bug_category | slice |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug25_green
红测分支: bug25_red
基线提交: 3412f5ca94b630b93e37abc683f3aea81389338d

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-025-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug25_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-025-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-025-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Snapshot alert data independently so later zone evaluations cannot mutate historical alert records.

## 验证命令（verify_cmds）
```bash
go test ./internal/geofence -count=1 -run TestEvaluateHistoryPreservesHistoricalAlerts
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：历史告警切片与本次评估结果复用底层数组，EvaluateHistory 使用 append(history[:0], alerts...) 覆盖历史记录。生产文件/符号：internal/geofence/service.go 的 Service.EvaluateHistory。调用链：区域评估 → 历史告警缓存 → append 复用数组。失效原因：后续评估改变历史告警，报表和审计记录失真。证据：红测发现 old[0].RuleID 从 old 被覆盖为 r1。 |

## 成功标准（success_criteria）
目标行为：新评估结果与历史告警完全独立。边界：空历史、容量足够、容量不足、多规则和重复评估均安全。合法场景：历史记录保持原值，新告警单独返回。验证标准：回归测试、全量测试和 race 测试通过。

## docker 打包验证
- bug25_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
