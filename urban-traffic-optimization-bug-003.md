# 出题包 BUG-003

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-003 |
| task_type | bugfix |
| bug_category | slice |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug3_green
红测分支: bug3_red
基线提交: cdcc1fa1929fd6871d3324c8133e488cb9942aa4

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-003-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug3_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-003-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-003-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
路网可达性分析返回结果后，不得继续污染原始道路列表。

## 验证命令（verify_cmds）
```bash
go test ./internal/road -count=1 -run TestReachableSnapshotReturnsIndependentPaths
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：路网可达查询复用包级 reachableSnapshot 的底层数组，并将新结果追加到旧 backing array 上。生产文件/符号：internal/road/service.go 的 Service.ReachableSnapshot。调用链：路网拓扑查询 → ReachableSnapshot → Reachable → 共享切片复用。失效原因：后一次查询重置长度但复用容量，覆盖前一次返回切片的元素，调用方保存的路径结果因此被篡改。证据：bug3_red 回归测试先查询 road-a 再查询 road-b，第一次结果从 [road-b road-c] 变为 [road-c road-c]。 |

## 成功标准（success_criteria）
目标行为：每次路网可达查询都返回独立结果，后续查询不能改变已返回路径。边界：空图、环路、单节点和多分支拓扑都必须保持结果隔离。合法场景：连续查询 road-a 与 road-b，第一次结果仍为 [road-b road-c]，第二次结果为 [road-c]。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug3_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



