# 出题包 BUG-014

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-014 |
| task_type | diagnosis |
| bug_category | error |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug14_green
红测分支: bug14_red
$16983f25beb6cfa858eb12b6995c088d962ed234e

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-014-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug14_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-014-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-014-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
天气观测字段非法时，系统必须拒绝数据而不是静默替换默认值。

## 验证命令（verify_cmds）
```bash
go test ./internal/weather -count=1 -run TestRecordWithDefaultSurfacesInvalidObservation
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：RecordWithDefault 捕获天气观测校验错误后构造 clear 默认观测并返回 nil，调用方无法知道原始数据被拒绝。生产文件/符号：internal/weather/service.go 的 Service.RecordWithDefault。调用链：天气数据接入 → RecordWithDefault → Record/Validate → 默认观测替换。失效原因：非法降雨、能见度或风速数据被伪装成正常天气，交通容量修正和告警结果因此失真。证据：bug14_red 使用 RainMM=-1 的观测，方法返回 nil 和空 ID 的 clear 观测。 |

## 成功标准（success_criteria）
目标行为：天气观测校验失败时必须返回原始 validation error，不得静默替换为 clear。边界：身份缺失、数值越界、正常观测和默认时间补全均需分别处理。合法场景：RainMM=-1 时调用方收到 rainfall is outside range。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug14_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



