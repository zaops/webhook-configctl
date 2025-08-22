# webhook-configctl

一个用于交互式管理 `webhook.yaml` 文件的命令行小工具。

如果这个项目对你有帮助，请考虑给个 ⭐ Star 或 🍴 Fork 支持一下！

## 使用方法

### 查看帮助信息

```bash
./webhook-configctl --help
```

### 添加 Webhook 配置

#### 交互式添加配置

启动交互式会话，通过友好的界面逐步配置 webhook：

```bash
./webhook-configctl add
```

功能特性：
- 📝 **智能提示**：自动补全文件路径和常用配置
- 🔧 **参数配置**：支持从请求体、请求头、URL 参数传递数据给脚本
- 🛡️ **安全规则**：内置 GitHub/GitLab 签名验证、IP 白名单等安全模板
- 📤 **输出控制**：可选择是否返回脚本执行结果
- 📁 **工作目录**：自定义脚本执行的工作目录

#### 生成配置模板

快速生成带中文注释的空白配置模板：

```bash
./webhook-configctl add --template
```

### 校验配置文件

检查 `webhook.yaml` 文件格式是否正确，验证必填字段和配置结构：

```bash
./webhook-configctl validate
```

校验内容：
- ✅ **格式检查**：YAML 语法正确性
- ✅ **字段验证**：必填字段完整性
- ✅ **结构校验**：配置项格式规范性

### 配置文件示例

生成的 `webhook.yaml` 文件结构：

```yaml
# webhook 配置文件
- id: deploy-app
  # 要执行的命令
  execute-command: /opt/deploy.sh
  # 命令工作目录
  command-working-directory: /opt
  # 是否将命令输出作为响应
  include-command-output-in-response: true
  # 传递给命令的参数
  pass-arguments-to-command:
    - name: branch
      source: payload
      envname: GIT_BRANCH
    - name: repository
      source: payload
      envname: REPO_NAME
  # 触发规则（GitHub 签名验证示例）
  trigger-rule:
    match:
      - type: payload-hash-sha1
        secret: your-secret-here
        parameter:
          source: header
          name: X-Hub-Signature
```

### 触发 Webhook

配置完成后，可通过 HTTP POST 请求触发 webhook：

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"branch": "main", "repository": "my-app"}' \
  http://your-server:9000/hooks/deploy-app
```