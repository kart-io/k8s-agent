package validation

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// ValidatePort 验证端口号是否在有效范围内 (1-65535)
func ValidatePort(port int, fieldName string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid %s port: %d (must be between 1-65535)", fieldName, port)
	}
	return nil
}

// ValidateHostPort 验证主机地址和端口组合
func ValidateHostPort(host string, port int, serviceName string) error {
	if host == "" {
		return fmt.Errorf("%s host is required", serviceName)
	}

	if err := ValidatePort(port, serviceName); err != nil {
		return err
	}

	return nil
}

// ValidateAddr 验证地址格式 (host:port)
// 这是基础版本，适用于大多数场景
func ValidateAddr(addr, fieldName string) error {
	if addr == "" {
		return fmt.Errorf("%s address is required", fieldName)
	}

	// 尝试解析地址
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid %s address format '%s': %v (expected host:port)", fieldName, addr, err)
	}

	// 允许空 host（表示监听所有接口）
	if host != "" {
		// 验证 host 是否为有效的 IP 或主机名
		if ip := net.ParseIP(host); ip == nil {
			// 不是 IP，检查是否为有效的主机名
			// 简单检查：不包含非法字符
			if strings.Contains(host, " ") {
				return fmt.Errorf("%s host '%s' contains invalid characters", fieldName, host)
			}
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid %s port '%s': must be a number", fieldName, portStr)
	}

	return ValidatePort(port, fieldName)
}

// ValidateListenAddr 验证监听地址格式 (:port 或 host:port)
// 允许空 host，适用于服务器监听地址
// 例如: ":8080", "0.0.0.0:8080", "127.0.0.1:8080"
func ValidateListenAddr(addr, fieldName string) error {
	if addr == "" {
		return fmt.Errorf("%s address is required", fieldName)
	}

	// 尝试解析地址
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid %s address format '%s': %v (expected :port or host:port)", fieldName, addr, err)
	}

	// host 可以为空（表示监听所有接口）
	if host != "" {
		// 验证是否为有效的 IP 地址
		if ip := net.ParseIP(host); ip == nil {
			return fmt.Errorf("%s host '%s' is not a valid IP address", fieldName, host)
		}
	}

	// 验证端口
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid %s port '%s': must be a number", fieldName, portStr)
	}

	return ValidatePort(port, fieldName)
}

// ValidatePositiveInt 验证正整数
func ValidatePositiveInt(value int, fieldName string) error {
	if value < 1 {
		return fmt.Errorf("%s must be > 0, got: %d", fieldName, value)
	}
	return nil
}

// ValidateNonNegativeInt 验证非负整数
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0, got: %d", fieldName, value)
	}
	return nil
}

// ValidatePositiveFloat 验证正浮点数
func ValidatePositiveFloat(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0, got: %f", fieldName, value)
	}
	return nil
}

// ValidateNonNegativeFloat 验证非负浮点数
func ValidateNonNegativeFloat(value float64, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0, got: %f", fieldName, value)
	}
	return nil
}

// ValidateFloatRange 验证浮点数是否在指定范围内
func ValidateFloatRange(value, min, max float64, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %f and %f, got: %f", fieldName, min, max, value)
	}
	return nil
}

// ValidateIntRange 验证整数是否在指定范围内
func ValidateIntRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d, got: %d", fieldName, min, max, value)
	}
	return nil
}

// ValidatePositiveDuration 验证正时长
func ValidatePositiveDuration(duration time.Duration, fieldName string) error {
	if duration <= 0 {
		return fmt.Errorf("%s must be > 0, got: %v", fieldName, duration)
	}
	return nil
}

// ValidateDurationRange 验证时长是否在指定范围内
func ValidateDurationRange(duration, min, max time.Duration, fieldName string) error {
	if duration < min || duration > max {
		return fmt.Errorf("%s must be between %v and %v, got: %v", fieldName, min, max, duration)
	}
	return nil
}

// ValidateRequired 验证必填字符串字段
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateEnum 验证枚举值
func ValidateEnum(value, fieldName string, allowedValues []string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid %s: %s (allowed values: %s)",
		fieldName, value, strings.Join(allowedValues, ", "))
}

// ValidateMaxValue 验证不超过最大值的整数
func ValidateMaxValue(value, max int, fieldName string) error {
	if value > max {
		return fmt.Errorf("%s must be <= %d, got: %d", fieldName, max, value)
	}
	return nil
}

// ValidateMinValue 验证不低于最小值的整数
func ValidateMinValue(value, min int, fieldName string) error {
	if value < min {
		return fmt.Errorf("%s must be >= %d, got: %d", fieldName, min, value)
	}
	return nil
}

// ValidateConnectionPool 验证连接池配置
// maxConns: 最大连接数
// minConns: 最小连接数 (可以是 idle connections)
func ValidateConnectionPool(maxConns, minConns int, serviceName string) error {
	if err := ValidateNonNegativeInt(maxConns, fmt.Sprintf("%s max_connections", serviceName)); err != nil {
		return err
	}

	if err := ValidateNonNegativeInt(minConns, fmt.Sprintf("%s min_connections", serviceName)); err != nil {
		return err
	}

	if minConns > maxConns {
		return fmt.Errorf("%s min_connections (%d) cannot be greater than max_connections (%d)",
			serviceName, minConns, maxConns)
	}

	return nil
}

// ValidateURL 验证 URL 格式 (简单版本,检查基本格式)
func ValidateURL(url, fieldName string) error {
	if url == "" {
		return fmt.Errorf("%s URL is required", fieldName)
	}

	// 检查基本的 URL 格式
	if !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "nats://") &&
		!strings.HasPrefix(url, "tcp://") {
		return fmt.Errorf("invalid %s URL: %s (must start with http://, https://, nats://, or tcp://)",
			fieldName, url)
	}

	return nil
}

// ValidateRedisDB 验证 Redis 数据库索引 (0-15)
func ValidateRedisDB(db int) error {
	return ValidateIntRange(db, 0, 15, "redis database index")
}

// ValidateTimeouts 验证超时配置的合理性
// 通常: dialTimeout < readTimeout < writeTimeout
func ValidateTimeouts(dial, read, write time.Duration, serviceName string) error {
	if dial <= 0 {
		return fmt.Errorf("%s dial_timeout must be > 0, got: %v", serviceName, dial)
	}

	if read <= 0 {
		return fmt.Errorf("%s read_timeout must be > 0, got: %v", serviceName, read)
	}

	if write <= 0 {
		return fmt.Errorf("%s write_timeout must be > 0, got: %v", serviceName, write)
	}

	return nil
}

// ValidateFileExists 验证文件是否存在
func ValidateFileExists(filePath, fieldName string) error {
	if filePath == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %s", fieldName, filePath)
		}
		return fmt.Errorf("failed to check %s: %w", fieldName, err)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file: %s", fieldName, filePath)
	}

	return nil
}

// ValidateDirExists 验证目录是否存在
func ValidateDirExists(dirPath, fieldName string) error {
	if dirPath == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s directory does not exist: %s", fieldName, dirPath)
		}
		return fmt.Errorf("failed to check %s directory: %w", fieldName, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory: %s", fieldName, dirPath)
	}

	return nil
}
