# 出题包 BUG-008

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-008 |
| task_type | bugfix |
| bug_category | slice |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug8_green
红测分支: bug8_red
基线提交: 18c7097b7770541f04d5e618af0ca8803677a9bd

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-008-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug8_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-008-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-008-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Copy signal phases deeply so changing movements in a copied phase does not mutate the original timing plan.

## 验证命令（verify_cmds）
```bash
go test ./internal/signal -count=1 -run TestCopyPlanDeepCopiesPhaseMovements
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：复制配时方案时只复制了 phases 外层切片，没有复制每个 SignalPhase.Movements 的底层数组。生产文件/符号：internal/signal/service.go 的 CopyPlan。调用链：方案复制 → CopyPlan → phases 浅拷贝 → movements 共享 backing array。失效原因：修改复制方案的放行方向会同步改变原方案，造成已审核方案配置被意外污染。证据：bug8_red 回归测试修改复制方案的第一个 movement 后，原方案值从 north-through 变成 east-through。 |

## 成功标准（success_criteria）
目标行为：复制配时方案必须深拷贝相位及其放行方向切片。边界：空 movements、多相位、多方向和重复修改都不能影响源方案。合法场景：修改副本 movements 后原方案仍保持原始放行方向。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug8_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
