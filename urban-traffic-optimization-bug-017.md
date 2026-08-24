# 出题包 BUG-017

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-017 |
| task_type | bugfix |
| bug_category | nil |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug17_green
红测分支: bug17_red
基线提交: c234c974e9bea8aba080a01006e0267b18aab179

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-017-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug17_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-017-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-017-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
设备查询结果为空时，安排巡检必须安全处理而不能解引用 nil。

## 验证命令（verify_cmds）
```bash
go test ./internal/maintenance -count=1 -run TestScheduleFromLookupHandlesMissingDevice
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：巡检安排入口直接解引用设备查询结果，缺失设备时传入 nil。生产文件/符号：internal/maintenance/service.go 的 Service.ScheduleFromLookup。调用链：设备查询 → ScheduleFromLookup → Schedule(*device)。失效原因：设备不存在不是合法 Device 值，直接解引用触发 nil pointer panic，调用方无法获得可处理结果。证据：bug17_red 的 TestScheduleFromLookupHandlesMissingDevice 传入 nil 后稳定复现 runtime error: invalid memory address or nil pointer dereference。 |

## 成功标准（success_criteria）
目标行为：设备查询为空时安排巡检不应 panic，并应返回明确的缺失设备结果。边界：nil 设备、正常设备、零时间和重复调用都需行为稳定。合法场景：设备存在时生成 scheduled 巡检并保留设备 ID；设备缺失时安全返回且不写入巡检记录。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug17_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
