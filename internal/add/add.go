package add

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Hook struct {
	ID                             string                 	`yaml:"id"`
	ExecuteCommand                 string                 	`yaml:"execute-command"`
	CommandWorkingDirectory        string                 	`yaml:"command-working-directory,omitempty"`
	IncludeCommandOutputInResponse bool                   	`yaml:"include-command-output-in-response"`
	PassArgumentsToCommand         []Argument             	`yaml:"pass-arguments-to-command,omitempty"`
	TriggerRule                    map[string]interface{} 	`yaml:"trigger-rule,omitempty"`
}

type Argument struct {
	Name    string 	`yaml:"name"`
	Source  string 	`yaml:"source"`
	EnvName string 	`yaml:"envname,omitempty"`
}

type Config struct {
	Hooks []Hook 	`yaml:"-"`
}

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "交互式添加 webhook 配置",
		Long:  "通过交互式界面添加新的 webhook 配置到 webhook.yaml 文件",
		RunE:  runAdd,
	}

	cmd.Flags().Bool("template", false, "生成带中文注释的空白 webhook.yaml 模板")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	template, _ := cmd.Flags().GetBool("template")

	if template {
		return generateTemplate()
	}

	return addHook()
}

func generateTemplate() error {
	template := `# webhook 配置文件
- id: demo
  # 要执行的命令
  execute-command: /opt/deploy.sh
  # 命令工作目录
  command-working-directory: /opt
  # 是否将命令输出作为响应
  include-command-output-in-response: true
  # 传递给命令的参数（可选）
  pass-arguments-to-command: []
  # 触发规则（可选）
  trigger-rule: {}
`

	return ioutil.WriteFile("webhook.yaml", []byte(template), 0644)
}

func addHook() error {
	hook := Hook{
		IncludeCommandOutputInResponse: true,
	}

	// Define icons
	icons := survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "❯"
		icons.Question.Format = "cyan"
		icons.Help.Text = "info"
		icons.Help.Format = "blue"
		icons.Error.Text = "✗"
		icons.Error.Format = "red"
		icons.SelectFocus.Text = "❯"
		icons.SelectFocus.Format = "green"
	})

	// Helper function for yes/no prompts
	askConfirm := func(message string, defaultVal bool) bool {
		var choice string
		defaultValue := "no"
		if defaultVal {
			defaultValue = "yes"
		}
		prompt := &survey.Select{
			Message: message,
			Options: []string{"yes", "no"},
			Default: defaultValue,
		}
		survey.AskOne(prompt, &choice, icons)
		return choice == "yes"
	}

	// 必填字段
	fmt.Println("📝 配置必填字段...")

	// 1. ID (必填)
	idPrompt := &survey.Input{
		Message: "请输入 ID:",
	}
	survey.AskOne(idPrompt, &hook.ID, survey.WithValidator(survey.Required), icons)

	// 2. execute-command (必填)
	cmdPrompt := &survey.Input{
		Message: "请输入执行命令路径:",
		Suggest: func(toComplete string) []string {
			return getPathSuggestions(toComplete)
		},
	}
	survey.AskOne(cmdPrompt, &hook.ExecuteCommand, survey.WithValidator(survey.Required), icons)

	fmt.Println("\n⚙️  配置可选字段...")

	// 3. command-working-directory
	fmt.Println("\n📁 配置工作目录")
	defaultDir := filepath.ToSlash(filepath.Dir(hook.ExecuteCommand))
	if askConfirm(fmt.Sprintf("是否使用自定义工作目录？(默认: %s)", defaultDir), false) {
		dirPrompt := &survey.Input{
			Message: "请输入命令工作目录:",
			Default: defaultDir,
			Help:    "脚本执行时的工作目录，影响相对路径的解析",
			Suggest: func(toComplete string) []string {
				return getPathSuggestions(toComplete)
			},
		}
		survey.AskOne(dirPrompt, &hook.CommandWorkingDirectory, icons)
	}

	// 4. include-command-output-in-response
	fmt.Println("\n📤 配置响应输出")
	selectPrompt := &survey.Select{
		Message: "是否将脚本输出返回给webhook调用方？",
		Options: []string{
			"true - 返回脚本输出(推荐用于调试)",
			"false - 不返回输出(适用于生产环境)",
		},
		Default: "true - 返回脚本输出(推荐用于调试)",
		Help:    "选择true可以看到脚本执行结果，便于调试问题",
	}
	var selected string
	survey.AskOne(selectPrompt, &selected, icons)
	hook.IncludeCommandOutputInResponse = strings.HasPrefix(selected, "true")

	// 5. pass-arguments-to-command
	fmt.Println("\n🔧 配置脚本参数")
	fmt.Println("   参数可以将webhook请求中的数据传递给你的脚本")
	fmt.Println("   常见用途：传递Git分支、仓库名、提交ID等信息")

	if askConfirm("是否需要向脚本传递参数？", false) {
		for {
			arg := Argument{}

			fmt.Println("\n📋 添加新参数:")

			namePrompt := &survey.Input{
				Message: "参数名称 (如: branch, repository, commit_id):",
				Help:    "这是参数的标识名，用于在脚本中引用",
			}
			survey.AskOne(namePrompt, &arg.Name, survey.WithValidator(survey.Required), icons)

			sourcePrompt := &survey.Select{
				Message: "请选择参数来源 (source):",
				Options: []string{
					"payload - 来自请求体JSON数据",
					"header - 来自HTTP请求头",
					"query - 来自URL查询参数",
					"url - 来自URL路径参数",
				},
			}
			var sourceSelected string
			survey.AskOne(sourcePrompt, &sourceSelected, icons)
			arg.Source = strings.Split(sourceSelected, " - ")[0]

			envPrompt := &survey.Input{
				Message: "环境变量名 (可选，如: GIT_BRANCH, REPO_NAME):",
				Help:    "参数值会以此环境变量名传递给脚本，留空则不设置环境变量",
			}
			survey.AskOne(envPrompt, &arg.EnvName, icons)

			hook.PassArgumentsToCommand = append(hook.PassArgumentsToCommand, arg)

			if !askConfirm("继续添加参数？", false) {
				break
			}
		}
	}

	// 6. trigger-rule
	fmt.Println("\n🛡️  配置安全规则")
	fmt.Println("   触发规则用于验证webhook请求的合法性")
	fmt.Println("   建议配置以提高安全性")

	if askConfirm("是否配置安全触发规则？", false) {
		templatePrompt := &survey.Select{
			Message: "请选择安全规则类型:",
			Options: []string{
				"ip-whitelist - IP白名单验证",
				"github-signature - GitHub签名验证",
				"gitlab-token - GitLab令牌验证",
				"自定义规则 - 高级用户自定义",
			},
		}
		var template string
			survey.AskOne(templatePrompt, &template, icons)

		ruleType := strings.Split(template, " - ")[0]

		switch ruleType {
		case "ip-whitelist":
			hook.TriggerRule = map[string]interface{}{
				"match": []map[string]interface{}{
					{
						"type":     "ip-whitelist",
						"ip-range": "192.168.1.0/24",
					},
				},
			}
		case "github-signature":
			hook.TriggerRule = map[string]interface{}{
				"match": []map[string]interface{}{
					{
						"type":   "payload-hash-sha1",
						"secret": "your-secret-here",
						"parameter": map[string]string{
							"source": "header",
							"name":   "X-Hub-Signature",
						},
					},
				},
			}
		case "gitlab-token":
			hook.TriggerRule = map[string]interface{}{
				"match": []map[string]interface{}{
					{
						"type":  "value",
						"value": "your-token-here",
						"parameter": map[string]string{
							"source": "header",
							"name":   "X-Gitlab-Token",
						},
					},
				},
			}
		case "自定义规则":
			jsonPrompt := &survey.Multiline{
				Message: "请输入自定义 JSON 规则:",
			}
			var jsonStr string
			survey.AskOne(jsonPrompt, &jsonStr, icons)

			// 简单解析 JSON 字符串为 map
			hook.TriggerRule = map[string]interface{}{
				"custom": jsonStr,
			}
		}
	}

	if err := appendToWebhookFile(hook); err != nil {
		return err
	}

	printSummary(hook)

	return nil
}

func printSummary(hook Hook) {
	fmt.Println("\n🎉 配置添加成功！")
	fmt.Println("以下是您新增的 webhook 配置摘要：")
	fmt.Println("----------------------------------------")
	fmt.Printf("ID: %s\n", hook.ID)
	fmt.Printf("执行命令: %s\n", hook.ExecuteCommand)
	if hook.CommandWorkingDirectory != "" {
		fmt.Printf("工作目录: %s\n", hook.CommandWorkingDirectory)
	}
	fmt.Printf("返回输出: %t\n", hook.IncludeCommandOutputInResponse)

	if len(hook.PassArgumentsToCommand) > 0 {
		fmt.Println("传递参数:")
		for _, arg := range hook.PassArgumentsToCommand {
			fmt.Printf("  - 名称: %s, 来源: %s", arg.Name, arg.Source)
			if arg.EnvName != "" {
				fmt.Printf(", 环境变量: %s", arg.EnvName)
			}
			fmt.Println()
		}
	}

	if len(hook.TriggerRule) > 0 {
		fmt.Println("安全规则: 已配置")
	}
	fmt.Println("----------------------------------------")
	fmt.Println("\n🚀 如何触发您的 Webhook:")
	fmt.Println("您可以通过向以下 URL 发送 POST 请求来触发此 webhook:")
	fmt.Printf("  http://<your-server-ip-or-domain>/hooks/%s\n\n", hook.ID)
	fmt.Println("例如，使用 curl:")
	curlCmd := fmt.Sprintf(`curl -X POST -H \"Content-Type: application/json\" \
  -d '{\"your_key\": \"your_value\"}' \
  http://<your-server-ip-or-domain>/hooks/%s`, hook.ID)
	fmt.Println(curlCmd)
	fmt.Println("\n请将 `<your-server-ip-or-domain>` 替换为您的 webhook 服务器的实际地址。")
}

func getPathSuggestions(toComplete string) []string {
	var suggestions []string

	dir := filepath.Dir(toComplete)
	if dir == "." {
		dir = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filepath.Base(toComplete)) {
			fullPath := filepath.ToSlash(filepath.Join(dir, name))
			if entry.IsDir() {
				if !strings.HasSuffix(fullPath, "/") {
					fullPath += "/"
				}
			}
			suggestions = append(suggestions, fullPath)
		}
	}

	return suggestions
}

func appendToWebhookFile(hook Hook) error {
	filename := "webhook.yaml"

	var existingHooks []Hook

	// 读取现有文件
	if data, err := ioutil.ReadFile(filename); err == nil {
		yaml.Unmarshal(data, &existingHooks)
	}

	// 添加新的 hook
	existingHooks = append(existingHooks, hook)

	// 生成带注释的 YAML
	var output strings.Builder
	output.WriteString("# webhook 配置文件\n")

	for i, h := range existingHooks {
		if i > 0 {
			output.WriteString("\n")
		}

		output.WriteString(fmt.Sprintf("- id: %s\n", h.ID))
		output.WriteString("  # 要执行的命令\n")
		output.WriteString(fmt.Sprintf("  execute-command: %s\n", h.ExecuteCommand))

		if h.CommandWorkingDirectory != "" {
			output.WriteString("  # 命令工作目录\n")
			output.WriteString(fmt.Sprintf("  command-working-directory: %s\n", h.CommandWorkingDirectory))
		}

		output.WriteString("  # 是否将命令输出作为响应\n")
		output.WriteString(fmt.Sprintf("  include-command-output-in-response: %t\n", h.IncludeCommandOutputInResponse))

		if len(h.PassArgumentsToCommand) > 0 {
			output.WriteString("  # 传递给命令的参数\n")
			output.WriteString("  pass-arguments-to-command:\n")
			for _, arg := range h.PassArgumentsToCommand {
				output.WriteString(fmt.Sprintf("    - name: %s\n", arg.Name))
				output.WriteString(fmt.Sprintf("      source: %s\n", arg.Source))
				if arg.EnvName != "" {
					output.WriteString(fmt.Sprintf("      envname: %s\n", arg.EnvName))
				}
			}
		} else {
			output.WriteString("  # 传递给命令的参数（可选）\n")
			output.WriteString("  pass-arguments-to-command: []\n")
		}

		if len(h.TriggerRule) > 0 {
			output.WriteString("  # 触发规则\n")
			ruleData, _ := yaml.Marshal(map[string]interface{}{"trigger-rule": h.TriggerRule})
			ruleLines := strings.Split(string(ruleData), "\n")
			for _, line := range ruleLines {
				if strings.TrimSpace(line) != "" {
					output.WriteString("  " + line + "\n")
				}
			}
		} else {
			output.WriteString("  # 触发规则（可选）\n")
			output.WriteString("  trigger-rule: {}\n")
		}
	}

	return ioutil.WriteFile(filename, []byte(output.String()), 0644)
}
