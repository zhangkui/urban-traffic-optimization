# 城市交通信号优化与路网分析平台

面向城市交通管理部门和交通研究机构的信号配时、流量采集、拥堵分析、路网仿真和优化建议平台。

## 已实现模块

- Go REST API：道路、路口、配时方案、流量、事件、仿真、优化候选方案和实时 SSE。
- 领域服务：道路拓扑、信号相位冲突校验、配时状态机、流量聚合、事件状态机、仿真执行、优化评分、设备心跳和传感器校验。
- Vue 3 + TypeScript：运营首页、道路/路口/配时/流量/事件列表、仿真任务、信号优化页面。
- 基础设施：Docker Compose、PostgreSQL/Redis、SQLite 本地开发存储、迁移和种子 SQL。

## 启动后端

```powershell
go run .
```

健康检查：`GET http://localhost:8080/healthz`

## 启动前端

```powershell
cd web
npm install
npm run dev
```

## Docker Compose

```powershell
docker compose up --build
```

## API 示例

```powershell
Invoke-RestMethod -Method Post http://localhost:8080/api/v1/roads -ContentType 'application/json' -Body '{"name":"人民路","code":"RM001","level":"arterial","speedLimit":50,"lanes":4,"lengthMeters":3200,"status":"active"}'
```

统一响应格式为 `{ "data": ... }`，失败响应为 `{ "error": { "code": ..., "message": ... } }`。
