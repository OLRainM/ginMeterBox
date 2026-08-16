package config

import (
	"encoding/json"
	"os"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Data     DataConfig     `json:"data"`
	Export   ExportConfig   `json:"export"`
	Report   ReportConfig   `json:"report"`
	Font     FontConfig     `json:"font"`
	Security SecurityConfig `json:"security"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type DataConfig struct {
	// DatabaseFile 是应用运行时使用的 SQLite 数据库文件。
	DatabaseFile string `json:"databaseFile"`

	// BillingFile 与 TotalMeterFile 仅用于一次性迁移和人工回滚；应用不会再以它们作为运行时数据源。
	BillingFile    string `json:"billingFile"`
	TotalMeterFile string `json:"totalMeterFile"`
}

type ExportConfig struct {
	Dir string `json:"dir"`
}

type ReportConfig struct {
	Dir string `json:"dir"`
}

type FontConfig struct {
	Bold    string `json:"bold"`
	Regular string `json:"regular"`
}

type SecurityConfig struct {
	AdminPasswordHash   string   `json:"adminPasswordHash"`
	SessionCookieSecure bool     `json:"sessionCookieSecure"`
	AllowedOrigins      []string `json:"allowedOrigins"`
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: ServerConfig{Port: ":8080"},
		Data: DataConfig{
			DatabaseFile:   "data/billing.db",
			BillingFile:    "data/billing_records.json",
			TotalMeterFile: "data/total_meter_records.json",
		},
		Export: ExportConfig{Dir: "exports"},
		Report: ReportConfig{Dir: "reports"},
		Font: FontConfig{
			Bold:    "C:\\Windows\\Fonts\\msyhbd.ttc",
			Regular: "C:\\Windows\\Fonts\\msyh.ttc",
		},
		Security: SecurityConfig{},
	}
}

// Load 从文件加载配置，文件不存在则使用默认值。
func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
