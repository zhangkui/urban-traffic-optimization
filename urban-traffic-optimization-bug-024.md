# 出题包 BUG-024

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-024 |
| task_type | diagnosis |
| bug_category | context |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug24_green
红测分支: bug24_red
基线提交: 4ce3554e364e3c31230d1c9a6f302d3ea78a4b0e

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-024-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug24_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-024-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-024-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Find why an expired priority request may still be granted by a delayed decision task.

## 验证命令（verify_cmds）
```bash
go test ./internal/priority -count=1 -run TestDecideWithContextRejectsCancelledRequest
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：带 context 的优先请求决策入口没有检查取消信号，直接调用 Decide。生产文件/符号：internal/priority/service.go 的 Service.DecideWithContext。调用链：优先请求 → 延迟决策 → DecideWithContext → Decide。失效原因：请求已取消时仍可能授予优先权。证据：红测使用已取消 context，结果 Granted=true。 |

## 成功标准（success_criteria）
目标行为：取消或过期请求不得授予优先权。边界：已取消、已过期、有效和未知请求均明确返回。合法场景：有效请求按等级授予。验证标准：诊断回归测试通过。

## docker 打包验证
- bug24_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
