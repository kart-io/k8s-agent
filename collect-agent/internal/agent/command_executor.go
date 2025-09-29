package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// CommandExecutor executes commands received from the central control plane
// It implements safety measures to ensure only read-only operations are performed
type CommandExecutor struct {
	clientset kubernetes.Interface
	clusterID string
	logger    *zap.Logger
	mu        sync.RWMutex

	// allowedTools defines which tools can be executed
	allowedTools map[string][]string
}

// NewCommandExecutor creates a new command executor with safety restrictions
func NewCommandExecutor(clientset kubernetes.Interface, clusterID string, logger *zap.Logger) *CommandExecutor {
	return &CommandExecutor{
		clientset: clientset,
		clusterID: clusterID,
		logger:    logger.With(zap.String("component", "command-executor")),
		allowedTools: map[string][]string{
			// kubectl read-only operations
			"kubectl": {
				"get", "describe", "logs", "top", "explain",
				"version", "cluster-info", "api-resources", "api-versions",
			},
			// System information commands
			"ps": {"aux", "-ef"},
			"df": {"-h"},
			"free": {"-h"},
			"uptime": {},
			"uname": {"-a"},
			"whoami": {},
			// Network diagnostics
			"ping": {"-c", "3"},
			"nslookup": {},
			"dig": {},
			"curl": {"-I", "-s", "--connect-timeout", "5"},
			"wget": {"--spider", "-T", "5"},
		},
	}
}

// Execute executes a command and returns the result
func (ce *CommandExecutor) Execute(ctx context.Context, cmd types.Command) *types.CommandResult {
	startTime := time.Now()

	result := &types.CommandResult{
		CommandID: cmd.ID,
		ClusterID: ce.clusterID,
		Status:    "success",
		Timestamp: startTime,
	}

	ce.logger.Info("Executing command",
		zap.String("command_id", cmd.ID),
		zap.String("tool", cmd.Tool),
		zap.String("action", cmd.Action),
		zap.Strings("args", cmd.Args))

	// Validate command safety
	if err := ce.validateCommand(cmd); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Command validation failed: %v", err)
		result.Duration = time.Since(startTime)
		ce.logger.Warn("Command validation failed",
			zap.String("command_id", cmd.ID),
			zap.Error(err))
		return result
	}

	// Execute based on command type
	switch cmd.Type {
	case "diagnostic":
		ce.executeDiagnosticCommand(ctx, cmd, result)
	case "info":
		ce.executeInfoCommand(ctx, cmd, result)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("Unknown command type: %s", cmd.Type)
	}

	result.Duration = time.Since(startTime)

	ce.logger.Info("Command execution completed",
		zap.String("command_id", cmd.ID),
		zap.String("status", result.Status),
		zap.Duration("duration", result.Duration))

	return result
}

// validateCommand ensures the command is safe to execute
func (ce *CommandExecutor) validateCommand(cmd types.Command) error {
	// Check if tool is allowed
	allowedActions, toolExists := ce.allowedTools[cmd.Tool]
	if !toolExists {
		return fmt.Errorf("tool '%s' is not allowed", cmd.Tool)
	}

	// For kubectl, perform additional validation
	if cmd.Tool == "kubectl" {
		return ce.validateKubectlCommand(cmd, allowedActions)
	}

	// For other tools, check if action is allowed
	if len(allowedActions) > 0 {
		actionAllowed := false
		for _, allowedAction := range allowedActions {
			if cmd.Action == allowedAction {
				actionAllowed = true
				break
			}
		}
		if !actionAllowed {
			return fmt.Errorf("action '%s' is not allowed for tool '%s'", cmd.Action, cmd.Tool)
		}
	}

	// Check for dangerous patterns in arguments
	if err := ce.checkArgumentSafety(cmd.Args); err != nil {
		return err
	}

	return nil
}

// validateKubectlCommand performs specific validation for kubectl commands
func (ce *CommandExecutor) validateKubectlCommand(cmd types.Command, allowedActions []string) error {
	// Check if the action is allowed
	actionAllowed := false
	for _, allowedAction := range allowedActions {
		if cmd.Action == allowedAction {
			actionAllowed = true
			break
		}
	}
	if !actionAllowed {
		return fmt.Errorf("kubectl action '%s' is not allowed", cmd.Action)
	}

	// Additional safety checks for specific kubectl commands
	switch cmd.Action {
	case "logs":
		// Logs command is safe but we might want to limit output
		for i, arg := range cmd.Args {
			if arg == "--follow" || arg == "-f" {
				return fmt.Errorf("following logs is not allowed for safety reasons")
			}
			// Limit tail lines
			if arg == "--tail" && i+1 < len(cmd.Args) {
				// This is acceptable for diagnostic purposes
				continue
			}
		}
	case "get", "describe":
		// These are read-only operations, generally safe
		// But we can add specific restrictions if needed
		break
	case "top":
		// Resource usage commands are safe
		break
	}

	return nil
}

// checkArgumentSafety checks if command arguments contain dangerous patterns
func (ce *CommandExecutor) checkArgumentSafety(args []string) error {
	dangerousPatterns := []string{
		"rm ", "delete", "destroy", "kill", "terminate",
		"sudo", "su", "chmod", "chown",
		"mount", "umount", "mkfs", "dd",
		"iptables", "ip route", "ifconfig",
		"systemctl", "service", "init.d",
		"crontab", "at ",
		"curl -X POST", "curl -X PUT", "curl -X DELETE",
		"wget -O", "wget --post",
		"ssh", "scp", "rsync",
		"docker run", "docker exec",
		"&&", "||", ";", "|", ">", ">>", "<",
		"$(", "`", "${",
	}

	argString := strings.Join(args, " ")
	argStringLower := strings.ToLower(argString)

	for _, pattern := range dangerousPatterns {
		if strings.Contains(argStringLower, strings.ToLower(pattern)) {
			return fmt.Errorf("dangerous pattern detected in arguments: '%s'", pattern)
		}
	}

	return nil
}

// executeDiagnosticCommand executes diagnostic commands
func (ce *CommandExecutor) executeDiagnosticCommand(ctx context.Context, cmd types.Command, result *types.CommandResult) {
	// Set timeout for command execution
	timeout := cmd.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // default timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	var execCmd *exec.Cmd
	if cmd.Tool == "kubectl" {
		// For kubectl, we need to build the full command
		args := append([]string{cmd.Action}, cmd.Args...)
		execCmd = exec.CommandContext(execCtx, "kubectl", args...)
	} else {
		// For other tools
		if cmd.Action != "" {
			args := append([]string{cmd.Action}, cmd.Args...)
			execCmd = exec.CommandContext(execCtx, cmd.Tool, args...)
		} else {
			execCmd = exec.CommandContext(execCtx, cmd.Tool, cmd.Args...)
		}
	}

	// Set environment variables if provided
	if len(cmd.Env) > 0 {
		env := execCmd.Environ()
		for key, value := range cmd.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		execCmd.Env = env
	}

	// Execute command
	output, err := execCmd.CombinedOutput()

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Status = "timeout"
			result.Error = "Command execution timed out"
		} else {
			result.Status = "failed"
			result.Error = err.Error()
		}
	}

	result.Output = string(output)

	// Limit output size to prevent excessive memory usage
	maxOutputSize := 1024 * 1024 // 1MB
	if len(result.Output) > maxOutputSize {
		result.Output = result.Output[:maxOutputSize] + "\n... (output truncated)"
	}
}

// executeInfoCommand executes information gathering commands
func (ce *CommandExecutor) executeInfoCommand(ctx context.Context, cmd types.Command, result *types.CommandResult) {
	// Info commands are similar to diagnostic commands but might have different handling
	ce.executeDiagnosticCommand(ctx, cmd, result)
}