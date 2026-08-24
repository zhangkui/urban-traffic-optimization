# 出题包 BUG-004

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-004 |
| task_type | diagnosis |
| bug_category | error |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug4_green
红测分支: bug4_red
基线提交: 4ddd8e917469e19ee7e19d3b5ae008d9bbf1bb26

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-004-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug4_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-004-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-004-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Find why a timing-plan approval failure loses its business error code and explain how the handler should preserve it.

## 验证命令（verify_cmds）
```bash
go test ./internal/compliance -count=1 -run TestAuthorizePublishPropagatesFailedReview
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：配时方案发布授权方法在 CanPublish 返回错误后错误地返回 nil，吞掉了审核失败结果。生产文件/符号：internal/compliance/service.go 的 Service.AuthorizePublish。调用链：方案发布 → AuthorizePublish → CanPublish → 审核结果检查。失效原因：失败分支没有继续返回原始 error，调用方误以为方案可发布。证据：bug4_red 上回归测试使用 Passed=false 且含 error finding 的审核结果，方法却返回 nil。 |

## 成功标准（success_criteria）
目标行为：未通过审核的配时方案必须阻止发布并返回明确错误。边界：缺少批准状态、审核失败、审核通过三种场景均需分别处理；不得吞掉底层错误。合法场景：approved 方案配合 Passed=false 的审核结果应返回 compliance review has blocking findings。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug4_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
