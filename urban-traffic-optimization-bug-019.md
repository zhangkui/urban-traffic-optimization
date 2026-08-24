# 出题包 BUG-019

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-019 |
| task_type | diagnosis |
| bug_category | slice |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug19_green
红测分支: bug19_red
基线提交: 218ec099b55918302ff7d0e8f4efeab11e78cdd6

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-019-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug19_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-019-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-019-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
合规结果排序不能改变调用方拥有的 accepted 切片顺序。

## 验证命令（verify_cmds）
```bash
go test ./internal/quality -count=1 -run TestSortAcceptedDoesNotMutateCallerSlice
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：排序直接操作 Result.Accepted 引用的调用方切片。生产文件/符号：internal/quality/service.go 的 SortAcceptedInPlace。调用链：合规校验结果 → 排序 → sort.Slice。失效原因：切片共享底层数组，排序改变外部顺序。证据：红测调用后 caller-owned slice 首元素从 a 变为 b。 |

## 成功标准（success_criteria）
目标行为：整理合规结果不得修改调用方拥有的原始切片。边界：空、单元素、重复时间和多元素均需安全。合法场景：结果排序正确且输入顺序保持。验证标准：诊断回归测试稳定失败并明确指出共享底层数组。

## docker 打包验证
- bug19_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
