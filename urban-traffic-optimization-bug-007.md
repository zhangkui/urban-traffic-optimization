# 出题包 BUG-007

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-007 |
| task_type | diagnosis |
| bug_category | nil |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug7_green
红测分支: bug7_red
基线提交: d9c33f2a6394eff32c18d780740498121f5fd6cb

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-007-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug7_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-007-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-007-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Find why optimization recommendation can dereference a missing baseline plan and define the expected empty-result behavior.

## 验证命令（verify_cmds）
```bash
go test ./internal/maintenance -count=1 -run TestLatestScoreWithoutInspectionReturnsZero
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：LatestScore 假设设备一定存在巡检记录，直接访问空 List 结果的第一个元素。生产文件/符号：internal/maintenance/service.go 的 Service.LatestScore。调用链：设备健康评分查询 → LatestScore → List → items[0]。失效原因：新设备或尚未安排巡检的设备返回空切片，直接索引触发 index out of range。证据：bug7_red 回归测试在无巡检记录设备上报告 runtime error: index out of range [0] with length 0。 |

## 成功标准（success_criteria）
目标行为：没有巡检记录时健康评分查询应返回安全默认值 0，不得崩溃。边界：无记录、单条记录、多条记录和指定状态筛选均需正确处理。合法场景：设备尚未安排巡检时 LatestScore 返回 0。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug7_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
