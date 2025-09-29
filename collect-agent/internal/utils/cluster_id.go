package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterIDDetector detects and generates a unique cluster ID
type ClusterIDDetector struct {
	clientset kubernetes.Interface
	logger    *zap.Logger
}

// NewClusterIDDetector creates a new cluster ID detector
func NewClusterIDDetector(clientset kubernetes.Interface, logger *zap.Logger) *ClusterIDDetector {
	return &ClusterIDDetector{
		clientset: clientset,
		logger:    logger.With(zap.String("component", "cluster-id-detector")),
	}
}

// DetectClusterID attempts to detect a unique identifier for the cluster
// It tries multiple methods in order of preference
func (d *ClusterIDDetector) DetectClusterID(ctx context.Context) (string, error) {
	d.logger.Info("Detecting cluster ID")

	// Method 1: Check environment variable
	if clusterID := os.Getenv("CLUSTER_ID"); clusterID != "" {
		d.logger.Info("Using cluster ID from environment variable", zap.String("cluster_id", clusterID))
		return clusterID, nil
	}

	// Method 2: Check if running in a cloud environment
	if clusterID, err := d.detectCloudClusterID(ctx); err == nil && clusterID != "" {
		d.logger.Info("Detected cloud cluster ID", zap.String("cluster_id", clusterID))
		return clusterID, nil
	}

	// Method 3: Use kube-system namespace UID
	if clusterID, err := d.detectFromKubeSystemUID(ctx); err == nil && clusterID != "" {
		d.logger.Info("Using kube-system namespace UID as cluster ID", zap.String("cluster_id", clusterID))
		return clusterID, nil
	}

	// Method 4: Generate from cluster info
	if clusterID, err := d.generateFromClusterInfo(ctx); err == nil && clusterID != "" {
		d.logger.Info("Generated cluster ID from cluster info", zap.String("cluster_id", clusterID))
		return clusterID, nil
	}

	// Method 5: Fallback - generate a random ID based on available info
	clusterID := d.generateFallbackID(ctx)
	d.logger.Warn("Using fallback cluster ID", zap.String("cluster_id", clusterID))
	return clusterID, nil
}

// detectCloudClusterID tries to detect cluster ID from cloud provider metadata
func (d *ClusterIDDetector) detectCloudClusterID(ctx context.Context) (string, error) {
	// Try to get cluster name from various cloud provider methods

	// AWS EKS: Check for EKS cluster name in ConfigMap
	if clusterID, err := d.checkAWSEKS(ctx); err == nil && clusterID != "" {
		return clusterID, nil
	}

	// GCP GKE: Check for GKE cluster name
	if clusterID, err := d.checkGoogleGKE(ctx); err == nil && clusterID != "" {
		return clusterID, nil
	}

	// Azure AKS: Check for AKS cluster name
	if clusterID, err := d.checkAzureAKS(ctx); err == nil && clusterID != "" {
		return clusterID, nil
	}

	return "", fmt.Errorf("no cloud cluster ID detected")
}

// checkAWSEKS checks for AWS EKS cluster information
func (d *ClusterIDDetector) checkAWSEKS(ctx context.Context) (string, error) {
	// Look for aws-auth ConfigMap which often contains cluster name
	cm, err := d.clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "aws-auth", metav1.GetOptions{})
	if err == nil && cm.Data != nil {
		// Try to extract cluster name from the configmap
		for key, value := range cm.Data {
			if strings.Contains(key, "mapRoles") || strings.Contains(key, "mapUsers") {
				// Parse the YAML content to find cluster ARN or name
				if clusterName := d.extractClusterNameFromAWSConfig(value); clusterName != "" {
					return clusterName, nil
				}
			}
		}
	}

	// Check for EKS-specific node labels
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil && len(nodes.Items) > 0 {
		node := nodes.Items[0]
		if clusterName, exists := node.Labels["alpha.eksctl.io/cluster-name"]; exists {
			return clusterName, nil
		}
		if clusterName, exists := node.Labels["eks.amazonaws.com/cluster-name"]; exists {
			return clusterName, nil
		}
	}

	return "", fmt.Errorf("EKS cluster name not found")
}

// checkGoogleGKE checks for Google GKE cluster information
func (d *ClusterIDDetector) checkGoogleGKE(ctx context.Context) (string, error) {
	// Check for GKE-specific node labels
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil && len(nodes.Items) > 0 {
		node := nodes.Items[0]

		// GKE nodes have cluster name in labels
		if clusterName, exists := node.Labels["cloud.google.com/gke-cluster-name"]; exists {
			return clusterName, nil
		}

		// Also check in provider ID
		if providerID := node.Spec.ProviderID; strings.Contains(providerID, "gce://") {
			// Extract project and zone info, but for cluster name we need labels
		}
	}

	return "", fmt.Errorf("GKE cluster name not found")
}

// checkAzureAKS checks for Azure AKS cluster information
func (d *ClusterIDDetector) checkAzureAKS(ctx context.Context) (string, error) {
	// Check for AKS-specific node labels
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil && len(nodes.Items) > 0 {
		node := nodes.Items[0]

		// AKS nodes have cluster name in labels
		if clusterName, exists := node.Labels["kubernetes.azure.com/cluster"]; exists {
			return clusterName, nil
		}

		// Check resource group and cluster info
		if resourceGroup, exists := node.Labels["kubernetes.azure.com/resource-group"]; exists {
			// The resource group often contains or is the cluster name
			return resourceGroup, nil
		}
	}

	return "", fmt.Errorf("AKS cluster name not found")
}

// detectFromKubeSystemUID uses the kube-system namespace UID as cluster identifier
func (d *ClusterIDDetector) detectFromKubeSystemUID(ctx context.Context) (string, error) {
	ns, err := d.clientset.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %w", err)
	}

	uid := string(ns.UID)
	if uid == "" {
		return "", fmt.Errorf("kube-system namespace UID is empty")
	}

	// Use first 12 characters of UID for shorter cluster ID
	if len(uid) > 12 {
		uid = uid[:12]
	}

	return fmt.Sprintf("k8s-%s", uid), nil
}

// generateFromClusterInfo generates cluster ID from various cluster information
func (d *ClusterIDDetector) generateFromClusterInfo(ctx context.Context) (string, error) {
	var info []string

	// Get server version
	version, err := d.clientset.Discovery().ServerVersion()
	if err == nil {
		info = append(info, version.GitVersion)
	}

	// Get first node name for uniqueness
	nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil && len(nodes.Items) > 0 {
		info = append(info, nodes.Items[0].Name)

		// Add creation timestamp for uniqueness
		if !nodes.Items[0].CreationTimestamp.IsZero() {
			info = append(info, nodes.Items[0].CreationTimestamp.Format("20060102"))
		}
	}

	// Get kube-system namespace creation time
	ns, err := d.clientset.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err == nil && !ns.CreationTimestamp.IsZero() {
		info = append(info, ns.CreationTimestamp.Format("20060102-150405"))
	}

	if len(info) == 0 {
		return "", fmt.Errorf("no cluster info available for ID generation")
	}

	// Create hash from combined info
	combined := strings.Join(info, "-")
	hash := sha256.Sum256([]byte(combined))
	shortHash := fmt.Sprintf("%x", hash)[:12]

	return fmt.Sprintf("k8s-%s", shortHash), nil
}

// generateFallbackID generates a fallback cluster ID
func (d *ClusterIDDetector) generateFallbackID(ctx context.Context) string {
	// Use current timestamp and hostname if available
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	// Create a hash from hostname and current time
	info := fmt.Sprintf("%s-%d", hostname, os.Getpid())
	hash := sha256.Sum256([]byte(info))
	shortHash := fmt.Sprintf("%x", hash)[:8]

	return fmt.Sprintf("fallback-%s", shortHash)
}

// extractClusterNameFromAWSConfig extracts cluster name from AWS auth config
func (d *ClusterIDDetector) extractClusterNameFromAWSConfig(config string) string {
	// This is a simplified parser - in production you might want to use a proper YAML parser
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for cluster ARN patterns
		if strings.Contains(line, "arn:aws:eks:") {
			// Extract cluster name from ARN like: arn:aws:eks:region:account:cluster/cluster-name
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				clusterName := parts[len(parts)-1]
				// Clean up any quotes or extra characters
				clusterName = strings.Trim(clusterName, `"'`)
				if clusterName != "" {
					return clusterName
				}
			}
		}
	}
	return ""
}