# 出题包 BUG-020

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-020 |
| task_type | bugfix |
| bug_category | context |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug20_green
红测分支: bug20_red
基线提交: c056b90f484106411f025441fa6c2e7e52930673

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-020-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug20_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-020-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-020-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Do not publish an expired forecast result after the request context has been cancelled.

## 验证命令（verify_cmds）
```bash
go test ./internal/forecast -count=1 -run TestPublishHonorsCancelledContext
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：预测发布入口接收取消上下文但完全忽略 Done 信号。生产文件/符号：internal/forecast/service.go 的 Service.Publish。调用链：预测任务 → Publish(ctx, result) → 发布结果。失效原因：请求已经取消时仍返回成功，过期预测可能进入实时数据流。证据：红测使用已取消 context 调用 Publish，结果未返回错误。 |

## 成功标准（success_criteria）
目标行为：上下文取消后不得发布预测结果并返回 context 错误。边界：未取消、已取消、截止时间到期和 nil 结果都需稳定。合法场景：有效上下文正常发布，取消上下文被拒绝。验证标准：回归测试和全量测试通过。

## docker 打包验证
- bug20_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
