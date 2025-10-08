package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset        *kubernetes.Clientset
	metricsClientset *versioned.Clientset
	config           *rest.Config
}

// NewClientFromKubeConfig 从 kubeconfig 创建客户端
func NewClientFromKubeConfig(kubeconfigData []byte) (*Client, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	return newClient(config)
}

// NewClientFromConfig 从 rest.Config 创建客户端
func NewClientFromConfig(config *rest.Config) (*Client, error) {
	return newClient(config)
}

// NewInClusterClient 创建 in-cluster 客户端
func NewInClusterClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	return newClient(config)
}

func newClient(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	metricsClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics clientset: %w", err)
	}

	return &Client{
		clientset:        clientset,
		metricsClientset: metricsClientset,
		config:           config,
	}, nil
}

// Clientset 返回 K8s 客户端
func (c *Client) Clientset() *kubernetes.Clientset {
	return c.clientset
}

// MetricsClientset 返回 metrics 客户端
func (c *Client) MetricsClientset() *versioned.Clientset {
	return c.metricsClientset
}

// CheckConnection 检查连接
func (c *Client) CheckConnection(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	return err
}

// GetServerVersion 获取服务器版本
func (c *Client) GetServerVersion(ctx context.Context) (string, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.GitVersion, nil
}
