package validate

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Hook struct {
	ID                             string                 `yaml:"id"`
	ExecuteCommand                 string                 `yaml:"execute-command"`
	CommandWorkingDirectory        string                 `yaml:"command-working-directory,omitempty"`
	IncludeCommandOutputInResponse bool                   `yaml:"include-command-output-in-response"`
	PassArgumentsToCommand         []Argument             `yaml:"pass-arguments-to-command,omitempty"`
	TriggerRule                    map[string]interface{} `yaml:"trigger-rule,omitempty"`
}

type Argument struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	EnvName string `yaml:"envname,omitempty"`
}

func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "校验 webhook.yaml 配置文件",
		Long:  "校验 webhook.yaml 文件是否符合官方格式要求",
		RunE:  runValidate,
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	filename := "webhook.yaml"

	// 读取文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("无法读取文件 %s: %v", filename, err)
	}

	// 解析 YAML
	var hooks []Hook
	if err := yaml.Unmarshal(data, &hooks); err != nil {
		return fmt.Errorf("YAML 格式错误: %v", err)
	}

	// 校验每个 hook
	for i, hook := range hooks {
		if err := validateHook(hook, i); err != nil {
			return fmt.Errorf("Hook #%d 校验失败: %v", i+1, err)
		}
	}

	fmt.Printf("✅ webhook.yaml 校验通过！共 %d 个 hook 配置\n", len(hooks))
	return nil
}

func validateHook(hook Hook, index int) error {
	// 校验必填字段
	if hook.ID == "" {
		return fmt.Errorf("缺少必填字段 'id'")
	}

	if hook.ExecuteCommand == "" {
		return fmt.Errorf("缺少必填字段 'execute-command'")
	}

	// 校验参数格式
	for j, arg := range hook.PassArgumentsToCommand {
		if arg.Name == "" {
			return fmt.Errorf("参数 #%d 缺少 'name' 字段", j+1)
		}
		if arg.Source == "" {
			return fmt.Errorf("参数 #%d 缺少 'source' 字段", j+1)
		}
	}

	fmt.Printf("✅ Hook '%s' 校验通过\n", hook.ID)
	return nil
}
