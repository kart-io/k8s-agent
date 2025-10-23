package config

import (
	"fmt"

	"github.com/spf13/pflag"
)

// PredictionOptions 预测配置选项
type PredictionOptions struct {
	TimeWindows      []string       `mapstructure:"time_windows" yaml:"time_windows" json:"time_windows"`
	AnomalyDetection AnomalyOptions `mapstructure:"anomaly_detection" yaml:"anomaly_detection" json:"anomaly_detection"`
}

// AnomalyOptions 异常检测配置
type AnomalyOptions struct {
	Contamination float64 `mapstructure:"contamination" yaml:"contamination" json:"contamination"`
	NEstimators   int     `mapstructure:"n_estimators" yaml:"n_estimators" json:"n_estimators"`
}

// NewPredictionOptions 创建默认的预测配置
func NewPredictionOptions() *PredictionOptions {
	return &PredictionOptions{
		TimeWindows: []string{"1h", "6h", "24h"},
		AnomalyDetection: AnomalyOptions{
			Contamination: 0.1,
			NEstimators:   100,
		},
	}
}

// Validate 验证配置
func (o *PredictionOptions) Validate() error {
	if o.AnomalyDetection.Contamination < 0 || o.AnomalyDetection.Contamination > 1 {
		return fmt.Errorf("contamination must be between 0 and 1")
	}
	if o.AnomalyDetection.NEstimators < 1 {
		return fmt.Errorf("n_estimators must be at least 1")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *PredictionOptions) AddFlags(fs *pflag.FlagSet) {
	fs.Float64Var(&o.AnomalyDetection.Contamination, "prediction.contamination", o.AnomalyDetection.Contamination, "Anomaly detection contamination")
	fs.IntVar(&o.AnomalyDetection.NEstimators, "prediction.n-estimators", o.AnomalyDetection.NEstimators, "Number of estimators")
}

// ApplyTo 将配置应用到目标接口
func (o *PredictionOptions) ApplyTo(target interface{}) error {
	return nil
}

// Complete 完成配置初始化
func (o *PredictionOptions) Complete() error {
	if len(o.TimeWindows) == 0 {
		o.TimeWindows = []string{"1h", "6h", "24h"}
	}
	if o.AnomalyDetection.Contamination <= 0 {
		o.AnomalyDetection.Contamination = 0.1
	}
	if o.AnomalyDetection.NEstimators <= 0 {
		o.AnomalyDetection.NEstimators = 100
	}
	return nil
}
