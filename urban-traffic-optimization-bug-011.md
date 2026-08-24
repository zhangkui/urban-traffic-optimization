# 出题包 BUG-011

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-011 |
| task_type | bugfix |
| bug_category | concurrency |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug11_green
红测分支: bug11_red
基线提交: 4fee8c29fa76158fcc9e71f3070814b09ba11166

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-011-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug11_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-011-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-011-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Prevent a stale heartbeat from changing a device back to online after an administrator disables it concurrently.

## 验证命令（verify_cmds）
```bash
go test ./internal/device -count=1 -run TestApplyHeartbeatDoesNotReviveDisabledDevice
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：设备心跳更新路径无条件把设备状态设置为 online，没有检查管理员已经设置的 disabled 状态。生产文件/符号：internal/device/service.go 的 Service.ApplyHeartbeat。调用链：控制器心跳上报 → ApplyHeartbeat → Get → 修改状态 → save。失效原因：管理员禁用设备与迟到心跳交错时，旧心跳覆盖控制面状态，设备重新出现在在线列表。证据：bug11_red 回归测试注册 disabled 设备后调用 ApplyHeartbeat，最终状态变为 online。 |

## 成功标准（success_criteria）
目标行为：管理员禁用的设备即使收到旧心跳也必须保持 disabled。边界：disabled、maintenance、offline 和正常在线设备的心跳处理要区分；并发交错时控制面状态优先。合法场景：disabled 设备收到心跳后状态仍为 disabled，同时记录心跳时间供审计。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug11_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
