package options

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// TracerOptions 分布式追踪配置选项
// 参考 OneX 项目的 JaegerOptions 设计,支持 OpenTelemetry 协议
type TracerOptions struct {
	// Enable 是否启用分布式追踪
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// Endpoint 追踪服务器地址 (OTLP gRPC endpoint)
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`

	// ServiceName 服务名称
	ServiceName string `mapstructure:"service_name" yaml:"service_name" json:"service_name"`

	// Environment 部署环境 (dev/test/staging/prod)
	Environment string `mapstructure:"environment" yaml:"environment" json:"environment"`

	// SampleRate 采样率 (0.0-1.0, 1.0表示100%采样)
	SampleRate float64 `mapstructure:"sample_rate" yaml:"sample_rate" json:"sample_rate"`

	// Insecure 是否使用非安全连接 (不使用TLS)
	Insecure bool `mapstructure:"insecure" yaml:"insecure" json:"insecure"`

	// Attributes 额外的资源属性
	Attributes map[string]string `mapstructure:"attributes" yaml:"attributes" json:"attributes"`
}

// NewTracerOptions 创建默认的追踪配置
func NewTracerOptions() *TracerOptions {
	return &TracerOptions{
		Enable:      false, // 默认关闭
		Endpoint:    "localhost:4317",
		ServiceName: "",
		Environment: "dev",
		SampleRate:  1.0, // 默认100%采样
		Insecure:    true,
		Attributes:  make(map[string]string),
	}
}

// Validate 验证配置
func (o *TracerOptions) Validate() error {
	if !o.Enable {
		return nil // 追踪是可选的,如果未启用则跳过验证
	}

	// 验证服务名称
	if err := validation.ValidateRequired(o.ServiceName, "tracer service name"); err != nil {
		return err
	}

	// 验证端点地址
	if err := validation.ValidateAddr(o.Endpoint, "tracer endpoint"); err != nil {
		return err
	}

	// 验证环境
	if err := validation.ValidateEnum(o.Environment, "tracer environment",
		[]string{"dev", "test", "staging", "prod"}); err != nil {
		return err
	}

	// 验证采样率
	if o.SampleRate < 0.0 || o.SampleRate > 1.0 {
		return fmt.Errorf("tracer sample_rate must be between 0.0 and 1.0, got: %f", o.SampleRate)
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *TracerOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.BoolVar(&o.Enable, prefix+"tracer.enable", o.Enable,
		"Enable distributed tracing")
	fs.StringVar(&o.Endpoint, prefix+"tracer.endpoint", o.Endpoint,
		"OTLP gRPC endpoint address (host:port)")
	fs.StringVar(&o.ServiceName, prefix+"tracer.service-name", o.ServiceName,
		"Service name for tracing resource")
	fs.StringVar(&o.Environment, prefix+"tracer.environment", o.Environment,
		"Deployment environment (dev/test/staging/prod)")
	fs.Float64Var(&o.SampleRate, prefix+"tracer.sample-rate", o.SampleRate,
		"Trace sampling rate (0.0-1.0, 1.0 = 100%)")
	fs.BoolVar(&o.Insecure, prefix+"tracer.insecure", o.Insecure,
		"Use insecure connection (no TLS)")
}

// Complete 完成配置初始化
func (o *TracerOptions) Complete() error {
	if !o.Enable {
		return nil
	}

	// 确保服务名称不为空
	if o.ServiceName == "" {
		return fmt.Errorf("tracer service_name is required when tracer is enabled")
	}

	// 确保环境有效
	if o.Environment == "" {
		o.Environment = "dev"
	}

	// 确保采样率在有效范围内
	if o.SampleRate < 0.0 {
		o.SampleRate = 0.0
	} else if o.SampleRate > 1.0 {
		o.SampleRate = 1.0
	}

	// 确保端点地址不为空
	if o.Endpoint == "" {
		o.Endpoint = "localhost:4317"
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *TracerOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"enable":      o.Enable,
				"endpoint":    o.Endpoint,
				"serviceName": o.ServiceName,
				"environment": o.Environment,
				"sampleRate":  o.SampleRate,
				"insecure":    o.Insecure,
				"attributes":  o.Attributes,
			},
		)
	}

	return nil
}

// SetupTracerProvider 设置全局 TracerProvider
// 这是一个便捷方法,用于初始化 OpenTelemetry TracerProvider
func (o *TracerOptions) SetupTracerProvider(ctx context.Context) error {
	if !o.Enable {
		return nil
	}

	// 创建 OTLP gRPC exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(o.Endpoint),
	}

	if o.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 创建资源
	attrs := []attribute.KeyValue{
		semconv.ServiceName(o.ServiceName),
		attribute.String("environment", o.Environment),
		attribute.String("exporter", "otlp"),
	}

	// 添加自定义属性
	for k, v := range o.Attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 TracerProvider
	bsp := tracesdk.NewBatchSpanProcessor(exporter)
	tp := tracesdk.NewTracerProvider(
		// 设置采样率
		tracesdk.WithSampler(tracesdk.ParentBased(
			tracesdk.TraceIDRatioBased(o.SampleRate),
		)),
		// 使用批处理
		tracesdk.WithSpanProcessor(bsp),
		// 设置资源
		tracesdk.WithResource(res),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	return nil
}

// WithTracerEnable 设置是否启用追踪
func WithTracerEnable(enable bool) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.Enable = enable
	}
}

// WithTracerEndpoint 设置追踪端点
func WithTracerEndpoint(endpoint string) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.Endpoint = endpoint
	}
}

// WithTracerServiceName 设置服务名称
func WithTracerServiceName(name string) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.ServiceName = name
	}
}

// WithTracerEnvironment 设置环境
func WithTracerEnvironment(env string) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.Environment = env
	}
}

// WithTracerSampleRate 设置采样率
func WithTracerSampleRate(rate float64) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.SampleRate = rate
	}
}

// WithTracerInsecure 设置是否使用非安全连接
func WithTracerInsecure(insecure bool) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.Insecure = insecure
	}
}

// WithTracerAttributes 设置额外属性
func WithTracerAttributes(attrs map[string]string) func(*TracerOptions) {
	return func(o *TracerOptions) {
		o.Attributes = attrs
	}
}

// WithTracerAttribute 添加单个属性
func WithTracerAttribute(key, value string) func(*TracerOptions) {
	return func(o *TracerOptions) {
		if o.Attributes == nil {
			o.Attributes = make(map[string]string)
		}
		o.Attributes[key] = value
	}
}
