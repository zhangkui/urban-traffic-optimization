# 出题包 BUG-018

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-018 |
| task_type | bugfix |
| bug_category | error |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug18_green
红测分支: bug18_red
$169a11f3c3d3b90d47a16ed5c1ff05f8fbfacafbb

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-018-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug18_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-018-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-018-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
事件响应计划中后续动作失败时，之前已启动的动作必须回滚。

## 验证命令（verify_cmds）
```bash
go test ./internal/response -count=1 -run TestApplyPlanRollsBackWhenOneActionFails
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：多动作响应计划逐个启动动作，后续动作失败时没有回滚已启动动作。生产文件/符号：internal/response/service.go 的 Service.ApplyPlan。调用链：事件响应计划 → Instantiate/Actions → ApplyPlan → Start。失效原因：错误路径直接返回，前序动作已变为 running，计划状态出现部分提交。证据：红测观察到 notify 状态从 pending 变为 running 后返回失败。 |

## 成功标准（success_criteria）
目标行为：任一动作启动失败时整个响应计划保持失败前状态。边界：零动作、单动作、多动作及中途失败均需稳定。合法场景：所有动作可启动时全部 running；任一动作冲突时已启动动作恢复 pending。验证标准：回归测试和全量 Go 测试通过。

## docker 打包验证
- bug18_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



