# deploy

> 对应 arch：`docs/arch/92-deployment.md`

## 文件清单（阶段 0 由 Codex 创建）

| 文件 | 用途 |
| --- | --- |
| `Dockerfile` | API 服务镜像（distroless） |
| `Dockerfile.worker` | Worker 镜像 |
| `docker-compose.dev.yml` | 本地依赖：postgres / redis / minio / asynqmon |
| `nginx/api.conf` | 反代 + 真实 IP（v1.1 #6） |
| `nginx/admin.conf` | 后台静态 + /api 反代 |
| `nginx/h5.conf` | C 端静态 + /api 反代 |
| `prometheus/rules.yml` | 告警规则（参考 arch 15） |

> Dockerfile 与 compose 模板已写在 `docs/arch/92-deployment.md`，按需拷贝到本目录。
