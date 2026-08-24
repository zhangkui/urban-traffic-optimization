# Urban Traffic Optimization — 25 Bug 题目设计

分支模型：orphan-redgreen
基线分支：main
绿色分支命名：bug<N>_green
红色验证分支：bug<N>_red（题目落位阶段创建）
题目总数：25
任务分配：15 bugfix、10 diagnosis
Bug 类型：concurrency 5、nil 5、slice 5、error 5、context 5

| 编号 | 任务 | 类型 | 业务域 | 设计目标 |
|---|---|---|---|---|
| 1 | bugfix | concurrency | 实时流量 | 并发采集同一路口时聚合计数丢失 |
| 2 | diagnosis | nil | 设备心跳 | 离线设备查询遇到缺失心跳记录 |
| 3 | bugfix | slice | 路网拓扑 | 拓扑邻接列表复用导致路径结果相互污染 |
| 4 | diagnosis | error | 配时审核 | 审核失败被包装后丢失业务错误码 |
| 5 | bugfix | context | 仿真任务 | 取消仿真后后台计算仍继续写入结果 |
| 6 | bugfix | concurrency | 事件告警 | 同一拥堵阈值并发触发重复事件 |
| 7 | diagnosis | nil | 优化候选 | 无基准方案时推荐逻辑解引用空对象 |
| 8 | bugfix | slice | 信号相位 | 相位复制遗漏 movements 深拷贝 |
| 9 | diagnosis | error | 报表导出 | CSV 写入错误被忽略导致导出假成功 |
| 10 | bugfix | context | SSE 大屏 | 客户端断开后订阅发送阻塞 |
| 11 | bugfix | concurrency | 设备状态 | 心跳与禁用操作产生状态倒退 |
| 12 | diagnosis | nil | 路口指标 | 无流量样本时平均指标访问空结果 |
| 13 | bugfix | slice | OD 需求矩阵 | 过滤结果修改原始矩阵切片 |
| 14 | diagnosis | error | 天气修正 | 天气数据校验错误被降级为默认天气 |
| 15 | bugfix | context | 后台任务 | 超时任务未释放队列占用资源 |
| 16 | diagnosis | concurrency | 干线协调 | 偏移方案更新与读取出现不一致 |
| 17 | bugfix | nil | 设备巡检 | 缺少设备对象时巡检调度访问字段崩溃 |
| 18 | bugfix | error | 事件联动 | 部分处置动作失败时整体状态错误更新 |
| 19 | diagnosis | slice | 合规审查 | findings 排序修改调用方共享切片 |
| 20 | bugfix | context | 实时预测 | 预测窗口取消后仍发布过期结果 |
| 21 | bugfix | concurrency | 自适应信号 | 相位决策并发应用覆盖新版本方案 |
| 22 | diagnosis | nil | 容量评估 | 缺失进口道配置时瓶颈计算访问空项 |
| 23 | bugfix | error | 排放报表 | 未知车型因子错误被静默跳过 |
| 24 | diagnosis | context | 优先控制 | 优先请求过期后决策任务仍继续授权 |
| 25 | bugfix | slice | 区域告警 | 告警快照复用底层切片导致历史记录变化 |

## 设计约束

- 每题跨越至少两个生产文件，关联真实交通业务流程。
- bugfix 题需要红测失败、修复后绿测通过；diagnosis 题只提交可复现红测和根因说明。
- 每题测试只进入对应 `bug<N>_green` 交付链；main 保留基线缺陷或未修复状态。
- 不使用空测试、恒真断言、占位实现或与交通业务无关的人工错误。
- 每题必须填写英文 `user_query`、中文 `gold_root_cause`、中文 `success_criteria` 和真实 Go 环境版本。
