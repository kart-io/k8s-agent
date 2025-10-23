package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// AnalysisOptions 分析配置选项
type AnalysisOptions struct {
	MinConfidence       float64 `mapstructure:"min_confidence" yaml:"min_confidence" json:"min_confidence"`
	MaxRecommendations  int     `mapstructure:"max_recommendations" yaml:"max_recommendations" json:"max_recommendations"`
	IncludeSimilarCases bool    `mapstructure:"include_similar_cases" yaml:"include_similar_cases" json:"include_similar_cases"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" yaml:"similarity_threshold" json:"similarity_threshold"`
	UseLLMFallback      bool    `mapstructure:"use_llm_fallback" yaml:"use_llm_fallback" json:"use_llm_fallback"` // 规则分析失败时使用LLM
}

// NewAnalysisOptions 创建默认的分析配置
func NewAnalysisOptions() *AnalysisOptions {
	return &AnalysisOptions{
		MinConfidence:       0.6,
		MaxRecommendations:  5,
		IncludeSimilarCases: true,
		SimilarityThreshold: 0.7,
		UseLLMFallback:      false,
	}
}

// Validate 验证配置
func (o *AnalysisOptions) Validate() error {
	if o.MinConfidence < 0 || o.MinConfidence > 1 {
		return fmt.Errorf("min_confidence must be between 0 and 1")
	}
	if o.MaxRecommendations < 1 {
		return fmt.Errorf("max_recommendations must be at least 1")
	}
	if o.SimilarityThreshold < 0 || o.SimilarityThreshold > 1 {
		return fmt.Errorf("similarity_threshold must be between 0 and 1")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *AnalysisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.Float64Var(&o.MinConfidence, "analysis.min-confidence", o.MinConfidence, "Minimum confidence threshold")
	fs.IntVar(&o.MaxRecommendations, "analysis.max-recommendations", o.MaxRecommendations, "Maximum number of recommendations")
	fs.BoolVar(&o.IncludeSimilarCases, "analysis.include-similar-cases", o.IncludeSimilarCases, "Include similar cases in analysis")
	fs.Float64Var(&o.SimilarityThreshold, "analysis.similarity-threshold", o.SimilarityThreshold, "Similarity threshold")
	fs.BoolVar(&o.UseLLMFallback, "analysis.use-llm-fallback", o.UseLLMFallback, "Use LLM fallback if rule-based analysis fails")
}

// ApplyTo 将配置应用到目标接口
func (o *AnalysisOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"minConfidence":       o.MinConfidence,
				"maxRecommendations":  o.MaxRecommendations,
				"includeSimilarCases": o.IncludeSimilarCases,
				"similarityThreshold": o.SimilarityThreshold,
				"useLLMFallback":      o.UseLLMFallback,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *AnalysisOptions) Complete() error {
	if o.MinConfidence <= 0 {
		o.MinConfidence = 0.6
	}
	if o.MaxRecommendations <= 0 {
		o.MaxRecommendations = 5
	}
	if o.SimilarityThreshold <= 0 {
		o.SimilarityThreshold = 0.7
	}
	return nil
}
