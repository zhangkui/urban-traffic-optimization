# 出题包 BUG-016

## 基本信息
| 字段 | 值 |
|---|---|
| session_id | <执行后填 UUID> |
| bug_id | BUG-016 |
| task_type | diagnosis |
| bug_category | concurrency |
| repro_determinism | deterministic |

## 仓库与环境（v2，repo_url）
分支模型: orphan-redgreen
绿测分支: bug16_green
红测分支: bug16_red
$1279ccfe68295834ec568368e0f514209add470df

| 字段 | 值 |
|---|---|
| repo_url | https://github.com/zhangkui/urban-traffic-optimization.git |
| base_branch | 见上方绿测分支/红测分支/基线提交四行字段 |
| agent workspace | `F:\go-bug-create\workspace\urban-traffic-optimization-bug-016-green` |
| go_version | go1.26.1 windows/amd64 (GOTOOLCHAIN=auto) |

#### 准备命令
```bash
git clone --single-branch -b bug16_green https://github.com/zhangkui/urban-traffic-optimization.git F:\go-bug-create\workspace/urban-traffic-optimization-bug-016-green
cd F:\go-bug-create\workspace/urban-traffic-optimization-bug-016-green
git remote remove origin
export GOTOOLCHAIN=local
go version
go build ./...
go test ./...
```

## 用户需求（user_query）
干线协调发布 offset 时，并发读取者必须看到完整一致的快照。

## 验证命令（verify_cmds）
```bash
go test -race ./internal/corridor -count=1 -run TestPublishedOffsetsConcurrentAccessIsRaceFree
```

## 参考答案（整理端质检用，严禁交给执行模型）
| 字段 | 值 |
|---|---|
| gold_root_cause | 中文根因：干线协调服务对 published offset map 进行逐项写入和无锁遍历读取，发布过程不是原子快照。生产文件/符号：internal/corridor/service.go 的 Service.PublishOffsets 与 PublishedOffsets。调用链：绿波方案发布 → published map 更新 → 实时协调读取。失效原因：读线程可能看到半更新方案，并发 map 访问触发 data race，严重时导致运行时崩溃。证据：bug16_red 的 -race 测试报告 PublishOffsets 与 PublishedOffsets 同时访问 map。 |

## 成功标准（success_criteria）
目标行为：offset 发布对读者呈现完整一致快照，并发读写不得产生 race。边界：空计划、单路口、多路口和连续发布都需安全；读者不能观察到半更新 map。合法场景：发布线程和读取线程各执行 100 次，所有读取均安全完成。验证标准：回归测试在 -race 下通过，go test ./...、go test -race ./... 和 go vet ./... 不失败。

## docker 打包验证
- bug16_green: amd64 not_run / arm64 not_run / container not_run.

## 轨迹收集清单
- collect 验收测试轨迹后才创建红测分支并落位回归测试。



