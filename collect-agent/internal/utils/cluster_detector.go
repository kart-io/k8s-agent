package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterIDDetector detects the cluster ID from various cloud providers
type ClusterIDDetector struct {
	clientset kubernetes.Interface
	logger    *zap.Logger
}

// NewClusterIDDetector creates a new cluster ID detector
func NewClusterIDDetector(clientset kubernetes.Interface, logger *zap.Logger) *ClusterIDDetector {
	return &ClusterIDDetector{
		clientset: clientset,
		logger:    logger.With(zap.String("component", "cluster-detector")),
	}
}

// DetectClusterID attempts to detect the cluster ID from various sources
func (d *ClusterIDDetector) DetectClusterID(ctx context.Context) (string, error) {
	d.logger.Info("Attempting to detect cluster ID")

	// Try different detection methods in order of reliability
	methods := []func(context.Context) (string, error){
		d.detectFromEnvironment,
		d.detectFromEKS,
		d.detectFromGKE,
		d.detectFromAKS,
		d.detectFromKubernetesUID,
		d.detectFromNodeLabels,
	}

	for i, method := range methods {
		clusterID, err := method(ctx)
		if err == nil && clusterID != "" {
			d.logger.Info("Cluster ID detected",
				zap.String("cluster_id", clusterID),
				zap.Int("method", i))
			return clusterID, nil
		}
		if err != nil {
			d.logger.Debug("Detection method failed",
				zap.Int("method", i),
				zap.Error(err))
		}
	}

	return "", fmt.Errorf("failed to detect cluster ID from any source")
}

// detectFromEnvironment checks for CLUSTER_ID environment variable
func (d *ClusterIDDetector) detectFromEnvironment(ctx context.Context) (string, error) {
	clusterID := os.Getenv("CLUSTER_ID")
	if clusterID == "" {
		return "", fmt.Errorf("CLUSTER_ID environment variable not set")
	}
	return clusterID, nil
}

// detectFromEKS detects cluster ID from AWS EKS
func (d *ClusterIDDetector) detectFromEKS(ctx context.Context) (string, error) {
	// In EKS, we can check the kube-system namespace for cluster information
	configMap, err := d.clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "aws-auth", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Check for EKS-specific annotations or labels
	if arn, ok := configMap.Annotations["eks.amazonaws.com/cluster-name"]; ok {
		return arn, nil
	}

	// Try to get cluster name from node provider ID
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return "", err
	}

	if len(nodes.Items) > 0 {
		providerID := nodes.Items[0].Spec.ProviderID
		// EKS provider ID format: aws:///us-west-2a/i-0123456789abcdef0
		if strings.HasPrefix(providerID, "aws://") {
			parts := strings.Split(providerID, "/")
			if len(parts) >= 4 {
				region := parts[3]
				// Try to construct cluster name from labels
				if clusterName, ok := nodes.Items[0].Labels["alpha.eksctl.io/cluster-name"]; ok {
					return fmt.Sprintf("eks-%s-%s", region, clusterName), nil
				}
				if clusterName, ok := nodes.Items[0].Labels["eks.amazonaws.com/cluster-name"]; ok {
					return fmt.Sprintf("eks-%s-%s", region, clusterName), nil
				}
			}
		}
	}

	return "", fmt.Errorf("not an EKS cluster or cluster ID not found")
}

// detectFromGKE detects cluster ID from Google GKE
func (d *ClusterIDDetector) detectFromGKE(ctx context.Context) (string, error) {
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return "", err
	}

	if len(nodes.Items) > 0 {
		node := nodes.Items[0]
		providerID := node.Spec.ProviderID

		// GKE provider ID format: gce://project-id/zone/instance-name
		if strings.HasPrefix(providerID, "gce://") {
			// Check for GKE-specific labels
			if clusterName, ok := node.Labels["cloud.google.com/gke-cluster-name"]; ok {
				parts := strings.Split(strings.TrimPrefix(providerID, "gce://"), "/")
				if len(parts) >= 2 {
					project := parts[0]
					zone := parts[1]
					return fmt.Sprintf("gke-%s-%s-%s", project, zone, clusterName), nil
				}
			}
		}
	}

	return "", fmt.Errorf("not a GKE cluster or cluster ID not found")
}

// detectFromAKS detects cluster ID from Azure AKS
func (d *ClusterIDDetector) detectFromAKS(ctx context.Context) (string, error) {
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return "", err
	}

	if len(nodes.Items) > 0 {
		node := nodes.Items[0]
		providerID := node.Spec.ProviderID

		// AKS provider ID format: azure:///subscriptions/xxx/resourceGroups/yyy/providers/Microsoft.Compute/virtualMachines/zzz
		if strings.HasPrefix(providerID, "azure://") {
			// Check for AKS-specific labels
			if clusterName, ok := node.Labels["kubernetes.azure.com/cluster"]; ok {
				// Extract resource group from provider ID
				parts := strings.Split(providerID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						resourceGroup := parts[i+1]
						return fmt.Sprintf("aks-%s-%s", resourceGroup, clusterName), nil
					}
				}
				return fmt.Sprintf("aks-%s", clusterName), nil
			}
		}
	}

	return "", fmt.Errorf("not an AKS cluster or cluster ID not found")
}

// detectFromKubernetesUID uses the kube-system namespace UID as cluster ID
func (d *ClusterIDDetector) detectFromKubernetesUID(ctx context.Context) (string, error) {
	namespace, err := d.clientset.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	uid := string(namespace.UID)
	if uid == "" {
		return "", fmt.Errorf("kube-system namespace UID is empty")
	}

	// Use a shortened version of the UID
	if len(uid) > 8 {
		return fmt.Sprintf("k8s-%s", uid[:8]), nil
	}
	return fmt.Sprintf("k8s-%s", uid), nil
}

// detectFromNodeLabels attempts to find cluster ID from node labels
func (d *ClusterIDDetector) detectFromNodeLabels(ctx context.Context) (string, error) {
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return "", err
	}

	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found in cluster")
	}

	node := nodes.Items[0]

	// Check common labels that might contain cluster identification
	commonLabels := []string{
		"cluster-id",
		"cluster-name",
		"kops.k8s.io/instancegroup",
		"node.kubernetes.io/instance-type",
	}

	for _, label := range commonLabels {
		if value, ok := node.Labels[label]; ok && value != "" {
			return fmt.Sprintf("custom-%s", value), nil
		}
	}

	// If all else fails, use the node name as a base
	if node.Name != "" {
		// Extract the first part of the node name (often contains cluster info)
		parts := strings.Split(node.Name, "-")
		if len(parts) > 0 {
			return fmt.Sprintf("node-%s", parts[0]), nil
		}
	}

	return "", fmt.Errorf("could not determine cluster ID from node labels")
}