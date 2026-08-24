# 出题包 BUG-023

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-023 |
| task_type | bugfix |
| bug_category | error |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug23_green
红测分支: bug23_red
$17baa7fca9653773ac23b0475f50fa214601941cf

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-023-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug23_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-023-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-023-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
遇到未知车型排放因子时，排放报表必须返回明确错误。

## 验证命令（verify_cmds）
```bash
go test ./internal/emission -count=1 -run TestEstimateStrictRejectsUnknownVehicleFactor
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：严格排放估算入口直接复用 Estimate，未知车型被静默跳过并返回 nil 错误。生产文件/符号：internal/emission/service.go 的 Service.EstimateStrict。调用链：排放报表 → EstimateStrict → Estimate → 因子查找。失效原因：数据缺少排放因子时报告看似成功但排放量不完整。证据：红测传入 motorcycle 因子后仍返回 nil error。 |

## 成功标准（success_criteria）
目标行为：未知车型或污染物因子立即返回明确错误。边界：空车型、未知车型、已知车型和混合车型均稳定。合法场景：所有因子存在时生成完整估算。验证标准：回归测试通过。

## docker 打包验证
- bug23_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



