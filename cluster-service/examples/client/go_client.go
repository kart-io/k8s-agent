package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// K8sClient K8s Agent 客户端 (查询参数风格)
type K8sClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewK8sClient 创建新的 K8s 客户端
func NewK8sClient(baseURL string) *K8sClient {
	return &K8sClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse 统一响应格式
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// buildURL 构建带查询参数的 URL
func (c *K8sClient) buildURL(path string, params map[string]string) string {
	u, _ := url.Parse(c.BaseURL + path)
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// doRequest 执行 HTTP 请求
func (c *K8sClient) doRequest(method, url string) (*APIResponse, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &apiResp, nil
}

// ===========================
// 集群管理 API
// ===========================

// ListClusters 列出所有集群
func (c *K8sClient) ListClusters(page, pageSize int) (*APIResponse, error) {
	params := map[string]string{
		"page":     fmt.Sprintf("%d", page),
		"pageSize": fmt.Sprintf("%d", pageSize),
	}
	url := c.buildURL("/api/k8s/clusters", params)
	return c.doRequest("GET", url)
}

// GetCluster 获取集群详情 (查询参数)
func (c *K8sClient) GetCluster(clusterID string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
	}
	url := c.buildURL("/api/k8s/cluster", params)
	return c.doRequest("GET", url)
}

// GetClusterHealth 获取集群健康状态 (查询参数)
func (c *K8sClient) GetClusterHealth(clusterID string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
	}
	url := c.buildURL("/api/k8s/cluster/health", params)
	return c.doRequest("GET", url)
}

// ===========================
// 命名空间管理 API
// ===========================

// ListNamespaces 列出命名空间 (查询参数)
func (c *K8sClient) ListNamespaces(clusterID string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
	}
	url := c.buildURL("/api/k8s/namespaces", params)
	return c.doRequest("GET", url)
}

// GetNamespace 获取命名空间详情 (查询参数)
func (c *K8sClient) GetNamespace(clusterID, namespace string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
	}
	url := c.buildURL("/api/k8s/namespace", params)
	return c.doRequest("GET", url)
}

// ===========================
// Pod 管理 API
// ===========================

// ListPods 列出 Pods (查询参数)
func (c *K8sClient) ListPods(clusterID, namespace string, page, pageSize int) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
		"page":      fmt.Sprintf("%d", page),
		"pageSize":  fmt.Sprintf("%d", pageSize),
	}
	url := c.buildURL("/api/k8s/pods", params)
	return c.doRequest("GET", url)
}

// GetPod 获取 Pod 详情 (查询参数)
func (c *K8sClient) GetPod(clusterID, namespace, podName string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
		"name":      podName,
	}
	url := c.buildURL("/api/k8s/pod", params)
	return c.doRequest("GET", url)
}

// GetPodLogs 获取 Pod 日志 (查询参数)
func (c *K8sClient) GetPodLogs(clusterID, namespace, podName, container string, tailLines int) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
		"name":      podName,
	}

	if container != "" {
		params["container"] = container
	}
	if tailLines > 0 {
		params["tailLines"] = fmt.Sprintf("%d", tailLines)
	}

	url := c.buildURL("/api/k8s/pod/logs", params)
	return c.doRequest("GET", url)
}

// ===========================
// Deployment 管理 API
// ===========================

// ListDeployments 列出 Deployments (查询参数)
func (c *K8sClient) ListDeployments(clusterID, namespace string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
	}
	url := c.buildURL("/api/k8s/deployments", params)
	return c.doRequest("GET", url)
}

// GetDeployment 获取 Deployment 详情 (查询参数)
func (c *K8sClient) GetDeployment(clusterID, namespace, deploymentName string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
		"name":      deploymentName,
	}
	url := c.buildURL("/api/k8s/deployment", params)
	return c.doRequest("GET", url)
}

// ===========================
// Node 管理 API
// ===========================

// ListNodes 列出 Nodes (查询参数)
func (c *K8sClient) ListNodes(clusterID string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
	}
	url := c.buildURL("/api/k8s/nodes", params)
	return c.doRequest("GET", url)
}

// GetNode 获取 Node 详情 (查询参数)
func (c *K8sClient) GetNode(clusterID, nodeName string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"name":      nodeName,
	}
	url := c.buildURL("/api/k8s/node", params)
	return c.doRequest("GET", url)
}

// ===========================
// Service 管理 API
// ===========================

// ListServices 列出 Services (查询参数)
func (c *K8sClient) ListServices(clusterID, namespace string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
	}
	url := c.buildURL("/api/k8s/services", params)
	return c.doRequest("GET", url)
}

// GetService 获取 Service 详情 (查询参数)
func (c *K8sClient) GetService(clusterID, namespace, serviceName string) (*APIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"namespace": namespace,
		"name":      serviceName,
	}
	url := c.buildURL("/api/k8s/service", params)
	return c.doRequest("GET", url)
}

// ===========================
// 使用示例
// ===========================

func main() {
	// 创建客户端
	client := NewK8sClient("http://localhost:8080")

	// 示例 1: 列出所有集群
	fmt.Println("=== 列出所有集群 ===")
	clusters, err := client.ListClusters(1, 20)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", clusters)
	}

	// 示例 2: 获取集群详情 (查询参数)
	fmt.Println("\n=== 获取集群详情 (查询参数) ===")
	cluster, err := client.GetCluster("cluster-123")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", cluster)
	}

	// 示例 3: 列出命名空间 (查询参数)
	fmt.Println("\n=== 列出命名空间 (查询参数) ===")
	namespaces, err := client.ListNamespaces("cluster-123")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", namespaces)
	}

	// 示例 4: 获取命名空间详情 (查询参数)
	fmt.Println("\n=== 获取命名空间详情 (查询参数) ===")
	namespace, err := client.GetNamespace("cluster-123", "default")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", namespace)
	}

	// 示例 5: 列出 Pods (查询参数)
	fmt.Println("\n=== 列出 Pods (查询参数) ===")
	pods, err := client.ListPods("cluster-123", "default", 1, 20)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", pods)
	}

	// 示例 6: 获取 Pod 详情 (查询参数)
	fmt.Println("\n=== 获取 Pod 详情 (查询参数) ===")
	pod, err := client.GetPod("cluster-123", "default", "my-pod")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", pod)
	}

	// 示例 7: 获取 Pod 日志 (查询参数)
	fmt.Println("\n=== 获取 Pod 日志 (查询参数) ===")
	logs, err := client.GetPodLogs("cluster-123", "default", "my-pod", "app", 100)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", logs)
	}

	// 示例 8: 列出 Deployments (查询参数)
	fmt.Println("\n=== 列出 Deployments (查询参数) ===")
	deployments, err := client.ListDeployments("cluster-123", "default")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", deployments)
	}

	// 示例 9: 获取 Deployment 详情 (查询参数)
	fmt.Println("\n=== 获取 Deployment 详情 (查询参数) ===")
	deployment, err := client.GetDeployment("cluster-123", "default", "my-deployment")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", deployment)
	}

	// 示例 10: 列出 Nodes (查询参数)
	fmt.Println("\n=== 列出 Nodes (查询参数) ===")
	nodes, err := client.ListNodes("cluster-123")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", nodes)
	}

	// 示例 11: 测试 URL 编码 - 包含特殊字符的命名空间
	fmt.Println("\n=== 测试 URL 编码 (kube-system) ===")
	kubeSystem, err := client.GetNamespace("cluster-123", "kube-system")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", kubeSystem)
	}
}
