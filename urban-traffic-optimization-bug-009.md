# 出题包 BUG-009

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-009 |
| task_type | diagnosis |
| bug_category | error |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug9_green
红测分支: bug9_red
基线提交: b17b242528c4465fecc716da3e0b5dd107c123b5

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-009-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug9_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-009-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-009-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
Find why a failed CSV writer can still produce a successful export response and identify the missing error propagation.

## 验证命令（verify_cmds）
```bash
go test ./internal/reporting -count=1 -run TestTrafficCSVUncheckedReturnsWriteError
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：CSV 导出方法忽略 csv.Writer.Write 返回值，也不读取 writer.Error，最终无条件返回 nil。生产文件/符号：internal/reporting/service.go 的 Service.TrafficCSVUnchecked。调用链：报表导出 → CSV writer → 底层 io.Writer → 错误返回。失效原因：磁盘、网络响应或导出目标不可写时，调用方收到成功结果，无法重试或提示用户。证据：bug9_red 使用始终返回错误的 writer，方法仍返回 nil。 |

## 成功标准（success_criteria）
目标行为：CSV 导出必须传播表头、数据行和 Flush 阶段的底层写入错误。边界：空数据、部分写入失败、Flush 失败和正常写入都要返回准确结果。合法场景：目标 writer 返回 report destination unavailable 时，导出方法返回同一错误。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug9_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
