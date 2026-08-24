# 出题包 BUG-005

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-005 |
| task_type | bugfix |
| bug_category | context |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug5_green
红测分支: bug5_red
基线提交: c33b3e31eaea43bf543abb1a2706c529f08490c6

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-005-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug5_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-005-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-005-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
仿真请求被取消后，正在运行的仿真任务必须及时停止。

## 验证命令（verify_cmds）
```bash
go test ./internal/simulation -count=1 -run TestRunIgnoringCancellationHonorsContext
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：仿真执行入口 RunIgnoringCancellation 接收到调用方 context 后却改用 context.Background 调用 Run。生产文件/符号：internal/simulation/service.go 的 Service.RunIgnoringCancellation。调用链：仿真任务执行 → RunIgnoringCancellation → Run → ctx.Done。失效原因：调用方取消信号没有传递到执行循环，任务继续完成，无法及时释放资源或更新 cancelled 状态。证据：bug5_red 使用已取消 context，回归结果为 error=nil 且 status=completed，而预期是 context.Canceled 和 cancelled。 |

## 成功标准（success_criteria）
目标行为：仿真执行必须尊重调用方 context 的取消和超时信号。边界：执行前取消、执行中取消、正常完成三种场景分别返回正确状态；取消时不得保存 completed 结果。合法场景：已排队任务收到 context.Canceled 后返回 context.Canceled，状态为 cancelled。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug5_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



