# 出题包 BUG-015

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-015 |
| task_type | bugfix |
| bug_category | context |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug15_green
红测分支: bug15_red
基线提交: 436bb72cdca4800c7f905d589b3422232ff25cda

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-015-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug15_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-015-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-015-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
后台交通任务超过截止时间后，必须通过 context 取消执行。

## 验证命令（verify_cmds）
```bash
go test ./internal/jobs -count=1 -run TestEnqueueWithContextReleasesTimedOutJob
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：EnqueueWithContext 接收了调用方的 deadline context，却在启动执行 goroutine 时改用 context.Background；executeWithContext 也没有使用传入 context。生产文件/符号：internal/jobs/queue.go 的 Queue.EnqueueWithContext/executeWithContext。调用链：后台任务入队 → context deadline → executeWithContext → handler。失效原因：超时信号无法到达 handler，任务持续 running，cancel map 条目和队列资源不能及时释放。证据：bug15_red 回归测试的 20ms context 超时后，任务在 100ms 内仍为 running。 |

## 成功标准（success_criteria）
目标行为：调用方 context 超时或取消后，后台 handler 必须收到信号，任务进入 cancelled 或 failed，并释放 cancel 槽位。边界：执行前超时、执行中超时、正常完成和显式 Cancel 均需正确处理。合法场景：20ms 超时的 slow job 在有限时间内结束且不再保持 running。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug15_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
