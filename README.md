# webhook-configctl

一个用于交互式管理 `webhook.yaml` 文件的命令行小工具。

## 安装

1.  克隆本项目。
2.  运行 `go build -o webhook-configctl.exe ./cmd/webhook-configctl` 来编译生成可执行文件。

## 使用方法

### 添加一个新的 Webhook 配置

此命令将启动一个交互式会话，帮助您添加一条新的 Webhook 到 `webhook.yaml` 文件中。

```bash
./webhook-configctl add
```

### 校验配置文件

此命令将检查当前目录下的 `webhook.yaml` 文件格式是否正确。

```bash
./webhook-configctl validate
```