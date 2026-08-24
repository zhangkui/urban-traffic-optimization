# 出题包 BUG-022

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-022 |
| task_type | diagnosis |
| bug_category | nil |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug22_green
红测分支: bug22_red
基线提交: dd78ebe73e4f2c576636e090aba492cc6a8c843b

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-022-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug22_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-022-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-022-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Find why capacity assessment can access a missing approach configuration and define the expected diagnostic result.

## 验证命令（verify_cmds）
```bash
go test ./internal/capacity -count=1 -run TestAssessConfiguredReportsMissingApproach
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：AssessConfigured 将缺失的 ApproachInput 指针直接解引用。生产文件/符号：internal/capacity/service.go 的 Service.AssessConfigured。调用链：路口配置查询 → AssessConfigured → Assess。失效原因：缺失进口道配置应转为诊断结果，却在解引用 nil 时 panic。证据：红测传入 nil 后复现 nil pointer panic。 |

## 成功标准（success_criteria）
目标行为：缺失配置返回可诊断结果而非 panic。边界：nil、单进口道、多进口道和默认配时均稳定。合法场景：合法配置正常计算容量，缺失配置明确标记问题。验证标准：诊断回归测试通过。

## docker 打包验证
- bug22_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
