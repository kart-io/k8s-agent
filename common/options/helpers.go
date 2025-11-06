package options

import (
	"fmt"
	"net"
	"reflect"

	"github.com/spf13/pflag"
)

// Completer 定义 Complete 接口
type Completer interface {
	Complete() error
}

// Validator 定义 Validate 接口
type Validator interface {
	Validate() error
}

// FlagAdder 定义 AddFlags 接口
type FlagAdder interface {
	AddFlags(*pflag.FlagSet)
}

// CompleteAll 对所有子选项调用 Complete() 方法
//
// 使用反射自动识别结构体中实现了 Completer 接口的字段，
// 并依次调用它们的 Complete() 方法。
//
// 使用示例：
//
//	func (o *ServerOptions) Complete() error {
//	    // 设置服务特定的默认值
//	    if o.Logging.InitialFields == nil {
//	        o.Logging.InitialFields = make(map[string]interface{})
//	    }
//	    o.Logging.InitialFields["service.name"] = "my-service"
//
//	    // 统一处理所有子选项的 Complete
//	    return CompleteAll(o)
//	}
func CompleteAll(opts interface{}) error {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("CompleteAll requires a struct, got %T", opts)
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过非导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 跳过 nil 指针
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// 检查字段是否实现了 Completer 接口
		if field.CanInterface() {
			if completer, ok := field.Interface().(Completer); ok {
				if err := completer.Complete(); err != nil {
					return fmt.Errorf("failed to complete %s: %w", fieldType.Name, err)
				}
			}
		}
	}

	return nil
}

// ValidateAll 对所有子选项调用 Validate() 方法并收集所有错误
//
// 使用反射自动识别结构体中实现了 Validator 接口的字段，
// 并依次调用它们的 Validate() 方法，收集所有错误。
//
// 使用示例：
//
//	func (o *ServerOptions) Validate() []error {
//	    // 可以添加服务特定的验证逻辑
//	    var errs []error
//
//	    // 统一验证所有子选项
//	    errs = append(errs, ValidateAll(o)...)
//
//	    return errs
//	}
func ValidateAll(opts interface{}) []error {
	var errs []error

	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return []error{fmt.Errorf("ValidateAll requires a struct, got %T", opts)}
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过非导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 跳过 nil 指针
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// 检查字段是否实现了 Validator 接口
		if field.CanInterface() {
			if validator, ok := field.Interface().(Validator); ok {
				if err := validator.Validate(); err != nil {
					errs = append(errs, fmt.Errorf("%s validation failed: %w", fieldType.Name, err))
				}
			}
		}
	}

	return errs
}

// AddFlagsAll 对所有子选项调用 AddFlags() 方法
//
// 使用反射自动识别结构体中实现了 FlagAdder 接口的字段，
// 并依次调用它们的 AddFlags() 方法。
//
// 使用示例：
//
//	func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
//	    // 统一添加所有子选项的 flags
//	    AddFlagsAll(o, fs)
//
//	    // 可以添加服务特定的 flags
//	    fs.StringVar(&o.CustomField, "custom", "", "Custom field")
//	}
func AddFlagsAll(opts interface{}, fs *pflag.FlagSet) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// 跳过非导出字段
		if !field.CanInterface() {
			continue
		}

		// 跳过 nil 指针
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// 检查字段是否实现了 FlagAdder 接口
		if adder, ok := field.Interface().(FlagAdder); ok {
			adder.AddFlags(fs)
		}
	}
}

// SetServiceName 设置日志中的服务名称
//
// 这是一个便捷函数，用于在 Complete() 方法中设置服务名称。
//
// 使用示例：
//
//	func (o *ServerOptions) Complete() error {
//	    SetServiceName(o.Logging, "my-service")
//	    return CompleteAll(o)
//	}
func SetServiceName(logging *LoggingOptions, serviceName string) {
	if logging == nil {
		return
	}

	if logging.InitialFields == nil {
		logging.InitialFields = make(map[string]interface{})
	}

	if _, ok := logging.InitialFields["service.name"]; !ok {
		logging.InitialFields["service.name"] = serviceName
	}
}

// CompleteWithServiceName 设置服务名称后调用 CompleteAll
//
// 这是一个最常用的便捷函数，组合了 SetServiceName 和 CompleteAll。
//
// 使用示例：
//
//	func (o *ServerOptions) Complete() error {
//	    return CompleteWithServiceName(o, o.Logging, "my-service")
//	}
func CompleteWithServiceName(opts interface{}, logging *LoggingOptions, serviceName string) error {
	SetServiceName(logging, serviceName)
	return CompleteAll(opts)
}

// === OneX 风格的工具函数 ===

// 定义单位常量 (字节大小)
const (
	_   = iota // ignore first value
	KiB = 1 << (10 * iota)
	MiB
	GiB
	TiB
	PiB
)

// Join 连接前缀字符串，用于构建嵌套的配置键
// 例如: Join("db", "mysql") -> "db.mysql."
func Join(prefixes ...string) string {
	if len(prefixes) == 0 {
		return ""
	}

	result := ""
	for _, p := range prefixes {
		if p != "" {
			if result != "" {
				result += "."
			}
			result += p
		}
	}

	if result != "" {
		result += "."
	}

	return result
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// CreateListener 创建网络监听器
// 参考 OneX 项目设计，用于创建 TCP 监听器并返回监听器和端口
// 适用于需要动态分配端口或验证监听地址的场景
//
// 参数:
//   - network: 网络类型（tcp, tcp4, tcp6）
//   - addr: 监听地址（:port 或 host:port）
//
// 返回:
//   - net.Listener: 监听器实例
//   - int: 实际监听的端口号
//   - error: 错误信息
//
// 使用示例:
//
//	ln, port, err := CreateListener("tcp", ":0")  // 动态分配端口
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ln.Close()
//	fmt.Printf("Listening on port %d\n", port)
func CreateListener(network, addr string) (net.Listener, int, error) {
	// 创建监听器
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %s %s: %w", network, addr, err)
	}

	// 获取实际端口
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()
		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return ln, tcpAddr.Port, nil
}

// GetFreePort 获取一个可用的空闲端口
// 通过创建临时监听器并立即关闭来获取可用端口
// 注意：在高并发场景下，端口可能在返回后被其他进程占用
//
// 返回:
//   - int: 可用的端口号
//   - error: 错误信息
//
// 使用示例:
//
//	port, err := GetFreePort()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Free port: %d\n", port)
func GetFreePort() (int, error) {
	ln, port, err := CreateListener("tcp", ":0")
	if err != nil {
		return 0, err
	}
	_ = ln.Close()
	return port, nil
}
