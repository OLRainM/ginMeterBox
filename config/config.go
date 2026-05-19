package config

import (
	"encoding/json"
	"os"
)

// Config 应用配置
type Config struct {
	Server ServerConfig `json:"server"`
	Data   DataConfig   `json:"data"`
	Export ExportConfig `json:"export"`
	Report ReportConfig `json:"report"`
	Font   FontConfig   `json:"font"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type DataConfig struct {
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

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: ServerConfig{Port: ":8080"},
		Data: DataConfig{
			BillingFile:    "data/billing_records.json",
			TotalMeterFile: "data/total_meter_records.json",
		},
		Export: ExportConfig{Dir: "exports"},
		Report: ReportConfig{Dir: "reports"},
		Font: FontConfig{
			Bold:    "C:\\Windows\\Fonts\\msyhbd.ttc",
			Regular: "C:\\Windows\\Fonts\\msyh.ttc",
		},
	}
}

// Load 从文件加载配置，文件不存在则使用默认值
func Load(path string) *Config {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, cfg)
	return cfg
}
