package yaml

// Config 表示API测试工具的统一配置
type Config struct {
	// 包含的其他配置文件路径
	Includes []string `yaml:"includes"`
	// API规范文件路径（单个文件，向后兼容）
	Spec string `yaml:"spec"`
	// API规范文件路径（多个文件）
	SpecFiles []string `yaml:"specFiles"`
	// API基础URL
	BaseURL string `yaml:"base_url"`
	// 输出目录
	OutputDir string `yaml:"output_dir"`
	// 超时时间（秒）
	Timeout int `yaml:"timeout"`
	// 是否详细输出
	Verbose bool `yaml:"verbose"`

	// 请求配置
	Request struct {
		// 请求头
		Headers map[string]string `yaml:"headers"`
		// 路径参数
		PathParams map[string]string `yaml:"path_params"`
		// 查询参数
		QueryParams map[string]string `yaml:"query_params"`
		// 请求体
		RequestBodies map[string]string `yaml:"request_bodies"`
	} `yaml:"request"`

	// 测试数据配置
	TestData struct {
		// 初始化脚本
		InitScript string `yaml:"init_script"`
		// 清理脚本
		CleanupScript string `yaml:"cleanup_script"`
		// 数据源
		Sources []DataSource `yaml:"sources"`
	} `yaml:"test_data"`

	// 测试场景配置
	Scenarios []Scenario `yaml:"scenarios"`

	// CI/CD 集成配置
	CI struct {
		// 输出格式
		OutputFormat string `yaml:"output_format"`
		// 失败阈值
		FailThreshold float64 `yaml:"fail_threshold"`
		// 通知配置
		Notifications struct {
			Slack string `yaml:"slack"`
			Email string `yaml:"email"`
		} `yaml:"notifications"`
	} `yaml:"ci"`

	// Web UI 配置
	WebUI struct {
		// 是否启用
		Enabled bool `yaml:"enabled"`
		// 端口
		Port int `yaml:"port"`
	} `yaml:"web_ui"`
}

// DataSource 表示测试数据源
type DataSource struct {
	// 数据源类型
	Type string `yaml:"type"`
	// 数据源路径
	Path string `yaml:"path"`
}

// Scenario 表示测试场景
type Scenario struct {
	// 场景名称
	Name string `yaml:"name"`
	// 场景描述
	Description string `yaml:"description"`
	// 测试步骤
	Steps []Step `yaml:"steps"`
}

// Step 表示测试步骤
type Step struct {
	// 步骤名称
	Name string `yaml:"name"`
	// 端点路径
	Endpoint string `yaml:"endpoint"`
	// HTTP方法
	Method string `yaml:"method"`
	// 请求体 - 支持字符串或对象格式
	RequestBody interface{} `yaml:"request_body"`
	// 提取变量
	Extract map[string]string `yaml:"extract"`
	// 依赖步骤
	Dependencies []string `yaml:"dependencies"`
	// 断言
	Assert map[string]interface{} `yaml:"assert"`
}
