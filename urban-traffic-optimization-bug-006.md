# 出题包 BUG-006

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-006 |
| task_type | bugfix |
| bug_category | concurrency |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug6_green
红测分支: bug6_red
$170ed5d964efd6ce02ce7f34d7a9ac0c2a3e757e7

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-006-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug6_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-006-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-006-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
多个网关同时上报时，批量更新设备状态必须保证并发安全。

## 验证命令（verify_cmds）
```bash
go test -race ./internal/device -count=1 -run TestEvaluateBatchConcurrentIsRaceFree
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：设备批量状态刷新由多个 goroutine 直接并发写入同一个普通 map，没有同步保护。生产文件/符号：internal/device/service.go 的 Service.EvaluateBatchConcurrent。调用链：设备状态刷新任务 → EvaluateBatchConcurrent → goroutine → result map。失效原因：并发 map 写会产生 data race，严重时触发 concurrent map writes。证据：bug6_red 的 -race 测试报告 service.go:148 的 mapassign_faststr 数据竞争。 |

## 成功标准（success_criteria）
目标行为：批量刷新设备状态时所有设备结果安全写入并完整返回。边界：空列表、单设备和多设备并发刷新都不得出现 race；结果数量必须与输入设备数量一致。合法场景：32 个设备同时刷新，返回 32 个状态结果。验证标准：回归测试在 -race 下通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug6_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



