package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// K8sNamePattern Kubernetes 资源名称的正则表达式
var K8sNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// K8sLabelPattern Kubernetes 标签的正则表达式
var K8sLabelPattern = regexp.MustCompile(`^[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?$`)

// ValidateK8sName 验证 Kubernetes 资源名称
func ValidateK8sName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("name must be no more than 253 characters")
	}
	if !K8sNamePattern.MatchString(name) {
		return fmt.Errorf("name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}
	return nil
}

// ValidateNamespace 验证命名空间名称
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	return ValidateK8sName(namespace)
}

// ValidateClusterID 验证集群 ID
func ValidateClusterID(clusterID string) error {
	if clusterID == "" {
		return fmt.Errorf("cluster ID cannot be empty")
	}
	if len(clusterID) > 253 {
		return fmt.Errorf("cluster ID must be no more than 253 characters")
	}
	return nil
}

// ValidateLabelKey 验证标签键
func ValidateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key cannot be empty")
	}

	// 检查是否包含前缀
	parts := strings.Split(key, "/")
	if len(parts) > 2 {
		return fmt.Errorf("label key can have at most one '/'")
	}

	// 验证前缀（如果存在）
	if len(parts) == 2 {
		prefix := parts[0]
		if len(prefix) > 253 {
			return fmt.Errorf("label key prefix must be no more than 253 characters")
		}
		// 前缀必须是域名格式
		if !isDomainName(prefix) {
			return fmt.Errorf("label key prefix must be a valid DNS subdomain")
		}
	}

	// 验证名称部分
	name := parts[len(parts)-1]
	if len(name) == 0 || len(name) > 63 {
		return fmt.Errorf("label key name must be 1-63 characters")
	}
	if !K8sLabelPattern.MatchString(name) {
		return fmt.Errorf("label key name must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character")
	}

	return nil
}

// ValidateLabelValue 验证标签值
func ValidateLabelValue(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("label value must be no more than 63 characters")
	}
	if value == "" {
		return nil // 空值是允许的
	}
	if !K8sLabelPattern.MatchString(value) {
		return fmt.Errorf("label value must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character")
	}
	return nil
}

// ValidateLabels 验证标签集合
func ValidateLabels(labels map[string]string) error {
	for key, value := range labels {
		if err := ValidateLabelKey(key); err != nil {
			return fmt.Errorf("invalid label key %q: %w", key, err)
		}
		if err := ValidateLabelValue(value); err != nil {
			return fmt.Errorf("invalid label value for key %q: %w", key, err)
		}
	}
	return nil
}

// ValidateAnnotations 验证注解集合
func ValidateAnnotations(annotations map[string]string) error {
	for key := range annotations {
		if err := ValidateLabelKey(key); err != nil {
			return fmt.Errorf("invalid annotation key %q: %w", key, err)
		}
		// 注解值没有长度限制，只要是合法的 UTF-8 字符串即可
	}
	return nil
}

// isDomainName 检查是否为有效的域名
func isDomainName(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}

	parts := strings.Split(s, ".")
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		if !K8sNamePattern.MatchString(part) {
			return false
		}
	}
	return true
}

// ValidateReplicas 验证副本数
func ValidateReplicas(replicas int32) error {
	if replicas < 0 {
		return fmt.Errorf("replicas must be non-negative")
	}
	if replicas > 10000 {
		return fmt.Errorf("replicas must not exceed 10000")
	}
	return nil
}

// ValidateContainerName 验证容器名称
func ValidateContainerName(name string) error {
	if name == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("container name must be no more than 253 characters")
	}
	// 容器名称规则与 K8s 资源名称相同
	if !K8sNamePattern.MatchString(name) {
		return fmt.Errorf("container name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}
	return nil
}

// ValidateImageName 验证镜像名称
func ValidateImageName(image string) error {
	if image == "" {
		return fmt.Errorf("image name cannot be empty")
	}
	// 简单验证：镜像名称应该包含至少一个字符
	if len(strings.TrimSpace(image)) == 0 {
		return fmt.Errorf("image name cannot be empty or whitespace")
	}
	return nil
}
