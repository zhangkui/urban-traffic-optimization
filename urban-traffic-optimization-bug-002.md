# 出题包 BUG-002

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-002 |
| task_type | diagnosis |
| bug_category | nil |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug2_green
红测分支: bug2_red
基线提交: 2d430b091280e62a274cb34fdfa1ec7b02b43e5b

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-002-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug2_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-002-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-002-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
设备没有心跳记录时，安排下一次巡检不得发生空指针崩溃。

## 验证命令（verify_cmds）
```bash
go test ./internal/device -count=1 -run TestStatusSummaryWithoutHeartbeatReturnsOffline
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：设备状态摘要在没有心跳记录时仍解引用 nil 的 heartbeat 指针。生产文件/符号：internal/device/service.go 的 Service.StatusSummary。调用链：设备状态查询 → StatusSummary → Get → heartbeat 时间差计算。失效原因：离线设备注册后 LastHeartbeat 为零值，heartbeat 指针保持 nil，now.Sub(*heartbeat) 触发 nil pointer dereference。证据：bug2_red 上的 TestStatusSummaryWithoutHeartbeatReturnsOffline 在运行时报告 invalid memory address or nil pointer dereference。 |

## 成功标准（success_criteria）
目标行为：没有心跳记录的已注册设备应安全返回 offline，而不是崩溃。边界：零值心跳、过期心跳、有效近期心跳和设备查询错误都必须得到确定结果。合法场景：新注册且状态为 offline 的检测器查询状态摘要。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug2_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
