# 出题包 BUG-013

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-013 |
| task_type | bugfix |
| bug_category | slice |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug13_green
红测分支: bug13_red
$17a540b7f4f1b46e61ed55ced28215b5033d2e8aa

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-013-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug13_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-013-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-013-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
筛选 OD 矩阵时，不得修改调用方持有的原始观测数据。

## 验证命令（verify_cmds）
```bash
go test ./internal/demand -count=1 -run TestFilterInPlaceDoesNotMutateStoredMatrix
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：OD 矩阵筛选路径通过 matrix.Pairs[:0] 复用原切片底层数组，并把过滤结果写回输入矩阵。生产文件/符号：internal/demand/service.go 的 Service.FilterInPlace。调用链：OD 报表筛选 → FilterInPlace → 原始 Pairs backing array → 矩阵存储对象。失效原因：报表筛选只是查询操作，却改变了已保存矩阵的 pair 数量和内容，后续分析得到不完整需求数据。证据：bug13_red 回归测试筛选 zone-a 后，原矩阵 Pairs 从 2 条变为 1 条。 |

## 成功标准（success_criteria）
目标行为：OD 矩阵筛选返回新矩阵，不修改输入矩阵及其 Pairs 切片。边界：空筛选、全匹配、部分匹配和重复筛选都必须保持源数据完整。合法场景：筛选 zone-a 后结果 1 条，原矩阵仍保留 2 条及原总量。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug13_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



