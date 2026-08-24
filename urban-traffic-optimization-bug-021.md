# 出题包 BUG-021

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-021 |
| task_type | bugfix |
| bug_category | concurrency |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug21_green
红测分支: bug21_red
基线提交: 66d4b17f18b26b1f5b496b1def794affd6693800

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-021-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug21_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-021-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-021-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
自适应信号并发更新时，较新的配时方案版本不能被旧写入覆盖。

## 验证命令（verify_cmds）
```bash
go test -race ./internal/adaptive -count=1 -run TestAdaptiveControllerConcurrentConfiguration
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：自适应控制器对 controllers map 的配置读写没有同步保护。生产文件/符号：internal/adaptive/service.go 的 Service.Configure 和 Get。调用链：自适应配置请求 → map 查找/写回 → 控制器读取。失效原因：并发 map 访问产生 data race，更新可能互相覆盖。证据：-race 测试报告 Configure 内 mapaccess 与 mapassign 并发冲突。 |

## 成功标准（success_criteria）
目标行为：并发配置和读取控制器时无 race 且版本更新一致。边界：单控制器、多控制器、重复配置和读取并发均安全。合法场景：多个配置请求完成后控制器保持合法配置。验证标准：-race 回归测试通过。

## docker 打包验证
- bug21_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。
