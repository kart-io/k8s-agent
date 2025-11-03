// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package health

import (
	"context"
	"fmt"
	"sync"
)

// Manager 健康检查管理器
// 管理多个健康检查器，支持并发检查
type Manager struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// NewManager 创建健康检查管理器
func NewManager() *Manager {
	return &Manager{
		checkers: make(map[string]Checker),
	}
}

// Register 注册健康检查器
func (m *Manager) Register(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[checker.Name()] = checker
}

// Unregister 注销健康检查器
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checkers, name)
}

// CheckResult 检查结果
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "healthy" or "unhealthy"
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// CheckAll 检查所有注册的检查器
func (m *Manager) CheckAll(ctx context.Context) ([]CheckResult, bool) {
	m.mu.RLock()
	checkers := make(map[string]Checker, len(m.checkers))
	for name, checker := range m.checkers {
		checkers[name] = checker
	}
	m.mu.RUnlock()

	results := make([]CheckResult, 0, len(checkers))
	allHealthy := true

	// 并发检查所有 checker
	var wg sync.WaitGroup
	resultChan := make(chan CheckResult, len(checkers))

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c Checker) {
			defer wg.Done()

			err := c.Check(ctx)
			result := CheckResult{Name: n}

			if err != nil {
				result.Status = "unhealthy"
				result.Error = err.Error()
			} else {
				result.Status = "healthy"
			}

			resultChan <- result
		}(name, checker)
	}

	// 等待所有检查完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for result := range resultChan {
		results = append(results, result)
		if result.Status != "healthy" {
			allHealthy = false
		}
	}

	return results, allHealthy
}

// Check 检查单个检查器
func (m *Manager) Check(ctx context.Context, name string) (*CheckResult, error) {
	m.mu.RLock()
	checker, ok := m.checkers[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("checker not found: %s", name)
	}

	result := &CheckResult{Name: name}
	err := checker.Check(ctx)

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	} else {
		result.Status = "healthy"
	}

	return result, nil
}
