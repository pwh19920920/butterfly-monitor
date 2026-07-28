# Dragonfly Monitor

Go 监控告警后台，基于 butterfly-admin DDD 架构二开。

## 技术栈

- Go 1.26 + Gin (butterfly) + GORM + MySQL
- JWT 鉴权 + 菜单权限
- VictoriaMetrics 时序
- Grafana 面板同步（当前 stub 可接 SDK）
- XXL-JOB 调度四个任务

## 分层

```
cmd → starter → {interfaces, application, infrastructure, job, config}
                      ↓
                 domain ← {types, common}
```

## 新增业务模块步骤

1. `internal/domain/entity/`
2. `internal/domain/repository/`
3. `internal/infrastructure/persistence/`
4. `internal/application/`
5. `internal/types/`
6. `internal/interfaces/` + starter 注册

## 运行

必须在 `backend/` 目录：

```bash
go run ./cmd
```

配置：`configs/config.yml`  
迁移：`migrations/dragonfly_monitor.sql`

## 约定

- 中文注释
- Handler `*_handler.go` / App `*_app.go` / Repo `*_repository.go`
- 实体表名 `t_*`
- BaseEntity 软删除
