# 出题包 BUG-010

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-010 |
| task_type | bugfix |
| bug_category | context |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug10_green
红测分支: bug10_red
基线提交: f816278aa3341a75b665195e35f75087837e8736

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-010-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug10_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-010-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-010-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
客户端断开后，实时 SSE 交通流发送不能永久阻塞。

## 验证命令（verify_cmds）
```bash
go test ./internal/realtime -count=1 -run TestSendToClientHonorsDisconnectedContext
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：SSE 客户端发送方法直接向无缓冲 client channel 写入，没有在发送操作中监听 HTTP 请求 context。生产文件/符号：internal/realtime/hub.go 的 Hub.SendToClient。调用链：实时大屏发布 → SSE 客户端发送 → client channel → HTTP 连接。失效原因：客户端断开或消费者停止读取时，发布 goroutine 永久阻塞，无法释放订阅资源。证据：bug10_red 使用已取消 context 和无人读取的 client channel，50ms 内发送未返回。 |

## 成功标准（success_criteria）
目标行为：SSE 客户端断开后，发送操作应立即停止并返回 context.Canceled。边界：已取消 context、正常接收、缓冲区满和请求结束等场景都不得永久阻塞。合法场景：客户端取消后向无消费者 channel 发送更新，方法在有限时间内返回取消错误。验证标准：回归测试通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug10_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
