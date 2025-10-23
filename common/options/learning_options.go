package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// LearningOptions 学习系统配置选项
type LearningOptions struct {
	EnableFeedback         bool   `mapstructure:"enable_feedback" yaml:"enable_feedback" json:"enable_feedback"`
	MinSamplesForAccuracy  int    `mapstructure:"min_samples_for_accuracy" yaml:"min_samples_for_accuracy" json:"min_samples_for_accuracy"`
	AccuracyUpdateInterval string `mapstructure:"accuracy_update_interval" yaml:"accuracy_update_interval" json:"accuracy_update_interval"`
	ExportLearningData     bool   `mapstructure:"export_learning_data" yaml:"export_learning_data" json:"export_learning_data"`
	ExportPath             string `mapstructure:"export_path" yaml:"export_path" json:"export_path"`
}

// NewLearningOptions 创建默认的学习配置
func NewLearningOptions() *LearningOptions {
	return &LearningOptions{
		EnableFeedback:         true,
		MinSamplesForAccuracy:  10,
		AccuracyUpdateInterval: "1h",
		ExportLearningData:     false,
		ExportPath:             "./data/learning",
	}
}

// Validate 验证配置
func (o *LearningOptions) Validate() error {
	if o.MinSamplesForAccuracy < 1 {
		return fmt.Errorf("min_samples_for_accuracy must be at least 1")
	}
	if o.ExportLearningData && o.ExportPath == "" {
		return fmt.Errorf("export_path is required when export_learning_data is enabled")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *LearningOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.EnableFeedback, "learning.enable-feedback", o.EnableFeedback, "Enable feedback collection")
	fs.IntVar(&o.MinSamplesForAccuracy, "learning.min-samples", o.MinSamplesForAccuracy, "Minimum samples for accuracy calculation")
	fs.StringVar(&o.AccuracyUpdateInterval, "learning.update-interval", o.AccuracyUpdateInterval, "Accuracy update interval")
	fs.BoolVar(&o.ExportLearningData, "learning.export-data", o.ExportLearningData, "Export learning data")
	fs.StringVar(&o.ExportPath, "learning.export-path", o.ExportPath, "Learning data export path")
}

// ApplyTo 将配置应用到目标接口
func (o *LearningOptions) ApplyTo(target interface{}) error {
	return nil
}

// Complete 完成配置初始化
func (o *LearningOptions) Complete() error {
	if o.MinSamplesForAccuracy <= 0 {
		o.MinSamplesForAccuracy = 10
	}
	if o.AccuracyUpdateInterval == "" {
		o.AccuracyUpdateInterval = "1h"
	}
	if o.ExportPath == "" {
		o.ExportPath = "./data/learning"
	}
	return nil
}
