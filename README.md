# Image Poller

通过 GitHub Actions 中转拉取 Docker Hub 镜像的命令行工具。

## 背景

由于网络隔离，无法直接访问 Docker Hub。本工具利用 GitHub Actions 作为中转，将 Docker Hub 镜像推送到 ghcr.io，然后从 ghcr.io 拉取到本地。

## 工作原理

1. 在 GitHub 仓库创建分支（分支名 = 镜像名）
2. GitHub Actions 自动触发，从 Docker Hub 拉取镜像并推送到 ghcr.io
3. 工具监控 Actions 执行状态
4. 执行成功后，从 ghcr.io 拉取镜像到本地
5. 重命名为原始镜像名

## 环境变量

| 变量名                    | 说明                           | 示例           |
|------------------------|------------------------------|--------------|
| `IMGPULL_GITHUB_TOKEN` | GitHub Personal Access Token | `ghp_xxxx`   |
| `IMGPULL_GITHUB_REPO`  | GitHub 仓库名                   | `owner/repo` |

## 配置存储

配置存储在 SQLite 数据库中（与镜像记录共用同一数据库）：
- Docker 连接模式（CLI/API）
- Docker API Host（仅 API 模式）

GitHub Token 和 Repo 继续使用环境变量，更安全。

## 安装

```bash
go build -o imgpull.exe .
```

## 使用

### 拉取镜像

```bash
# 拉取 latest 标签
imgpull pull nginx

# 拉取指定标签
imgpull pull nginx:1.21
imgpull pull prom/prometheus:v2.45.0

# 强制重新拉取
imgpull pull nginx --force
```

### 配置 Docker 连接

```bash
# 使用 Docker CLI（默认）
imgpull config docker --cli

# 使用 Docker API
imgpull config docker --api
imgpull config docker --api --host tcp://localhost:2375
```

### 查看历史

```bash
imgpull list
imgpull list -n 50  # 限制显示数量
```

### 清理资源

```bash
# 预览将要删除的资源（不实际删除）
imgpull prune --dry-run

# 删除所有分支（保留 master/main）和所有容器包
imgpull prune
```

## GitHub Actions 配置

仓库需要包含以下两个 workflow 文件：

### trans-image.yml

监听分支创建，自动拉取 latest 镜像：

```yaml
name: trans-on-create-branch
on:
  create

env:
  REGISTRY: ghcr.io

jobs:
  pull_and_push:
    runs-on: ubuntu-latest
    steps:
      - name: Pull Docker image
        run: docker pull ${{ github.ref_name }}
      - name: Log in to the Container registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: reTag and push Docker image
        run: |
          docker tag ${{ github.ref_name }} ${REGISTRY}/${{ github.actor }}/${{ github.ref_name }}
          docker push ${REGISTRY}/${{ github.actor }}/${{ github.ref_name }}
```

### pull-image-with-tag.yml

手动触发，拉取指定标签镜像：

```yaml
name: pull-image-with-tag
on:
  workflow_dispatch:
    inputs:
      tag:
        description: 'Will pull image tag'
        required: true
        default: 'latest'
        type: string

env:
  REGISTRY: ghcr.io

jobs:
  pull_and_push:
    runs-on: ubuntu-latest
    steps:
      - name: Pull Docker image
        run: docker pull ${{ github.ref_name }}:${{ inputs.tag }}
      - name: Log in to the Container registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: reTag and push Docker image
        run: |
          docker tag ${{ github.ref_name }}:${{ inputs.tag }} ${REGISTRY}/${{ github.actor }}/${{ github.ref_name }}:${{ inputs.tag }}
          docker push ${REGISTRY}/${{ github.actor }}/${{ github.ref_name }}:${{ inputs.tag }}
```

## Workflow 选择逻辑

| 场景                        | 使用的 Workflow              |
|---------------------------|---------------------------|
| latest 标签 + 新分支           | `trans-image.yml`（自动触发）   |
| latest 标签 + 已存在分支 + force | `pull-image-with-tag.yml` |
| 非 latest 标签               | `pull-image-with-tag.yml` |

## 项目结构

```
image-poller/
├── main.go                  # 入口
├── cmd/
│   ├── root.go              # rootCmd
│   ├── pull.go              # pullCmd
│   ├── list.go              # listCmd
│   ├── config.go            # configCmd
│   ├── prune.go             # pruneCmd
├── internal/
│   ├── config/
│   │   ├── config.go        # Config 结构、数据库读写适配
│   │   ├── env.go           # 环境变量
│   │   ├── path.go          # 配置路径、数据库路径
│   ├── db/
│   │   ├── db.go            # 数据库连接
│   │   ├── record.go        # ImageRecord 和操作方法
│   │   ├── config.go        # ConfigItem 和配置读写方法
│   ├── docker/
│   │   ├── client.go        # DockerClient 接口、NewClient
│   │   ├── cli.go           # CLIClient 实现
│   │   ├── api.go           # APIClient 实现
│   ├── github/
│   │   ├── client.go        # Client 结构、接口定义
│   │   ├── branch.go        # 分支操作
│   │   ├── workflow.go      # Workflow 操作
│   │   ├── package.go       # Package 操作
├── pkg/
│   ├── image/
│   │   ├── reference.go     # 镜像引用解析
│   ├── display/
│   │   ├── colors.go        # 颜色定义
│   │   ├── spinner.go       # Spinner 动画
│   │   ├── tty.go           # TTY 输出
│   │   ├── event.go         # Event 处理
├── utils/
│   ├── stopwatch.go         # 耗时记录
│   ├── retry.go             # 重试逻辑
├── go.mod
└── README.md
```

## 技术栈

- Go 1.21+
- GitHub API: `github.com/google/go-github/v85`
- Docker SDK: `github.com/moby/moby`
- CLI: `github.com/spf13/cobra`
- 数据库: GORM + `github.com/glebarez/sqlite`（纯 Go SQLite）