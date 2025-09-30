package utils

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDetectFromEnvironment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	clientset := fake.NewSimpleClientset()
	detector := NewClusterIDDetector(clientset, logger)

	// Test with environment variable set
	expectedID := "test-cluster-123"
	os.Setenv("CLUSTER_ID", expectedID)
	defer os.Unsetenv("CLUSTER_ID")

	clusterID, err := detector.detectFromEnvironment(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if clusterID != expectedID {
		t.Errorf("Expected cluster ID %s, got %s", expectedID, clusterID)
	}

	// Test without environment variable
	os.Unsetenv("CLUSTER_ID")
	_, err = detector.detectFromEnvironment(context.Background())
	if err == nil {
		t.Error("Expected error when CLUSTER_ID not set, got nil")
	}
}

func TestDetectFromKubernetesUID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a fake namespace with UID
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID("12345678-1234-1234-1234-123456789012"),
		},
	}

	clientset := fake.NewSimpleClientset(namespace)
	detector := NewClusterIDDetector(clientset, logger)

	clusterID, err := detector.detectFromKubernetesUID(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedPrefix := "k8s-12345678"
	if clusterID != expectedPrefix {
		t.Errorf("Expected cluster ID to start with %s, got %s", expectedPrefix, clusterID)
	}
}

func TestDetectFromEKS(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a fake EKS node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-10-0-1-100.us-west-2.compute.internal",
			Labels: map[string]string{
				"eks.amazonaws.com/cluster-name": "my-eks-cluster",
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: "aws:///us-west-2a/i-0123456789abcdef0",
		},
	}

	clientset := fake.NewSimpleClientset(node)
	detector := NewClusterIDDetector(clientset, logger)

	clusterID, err := detector.detectFromEKS(context.Background())
	if err != nil {
		t.Logf("EKS detection returned error (expected in fake env): %v", err)
	} else {
		t.Logf("Detected EKS cluster ID: %s", clusterID)
	}
}

func TestDetectFromGKE(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a fake GKE node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gke-cluster-1-default-pool-12345678-abcd",
			Labels: map[string]string{
				"cloud.google.com/gke-cluster-name": "cluster-1",
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: "gce://my-project/us-central1-a/gke-cluster-1-default-pool-12345678-abcd",
		},
	}

	clientset := fake.NewSimpleClientset(node)
	detector := NewClusterIDDetector(clientset, logger)

	clusterID, err := detector.detectFromGKE(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedID := "gke-my-project-us-central1-a-cluster-1"
	if clusterID != expectedID {
		t.Errorf("Expected cluster ID %s, got %s", expectedID, clusterID)
	}
}

func TestDetectFromNodeLabels(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a fake node with custom labels
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-node-1",
			Labels: map[string]string{
				"cluster-name": "production-cluster",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node)
	detector := NewClusterIDDetector(clientset, logger)

	clusterID, err := detector.detectFromNodeLabels(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if clusterID == "" {
		t.Error("Expected non-empty cluster ID")
	}
	t.Logf("Detected cluster ID from node labels: %s", clusterID)
}

func TestDetectClusterID(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Test with environment variable (highest priority)
	expectedID := "env-cluster-123"
	os.Setenv("CLUSTER_ID", expectedID)
	defer os.Unsetenv("CLUSTER_ID")

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID("12345678-1234-1234-1234-123456789012"),
		},
	}

	clientset := fake.NewSimpleClientset(namespace)
	detector := NewClusterIDDetector(clientset, logger)

	clusterID, err := detector.DetectClusterID(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if clusterID != expectedID {
		t.Errorf("Expected cluster ID %s, got %s", expectedID, clusterID)
	}

	// Test without environment variable (fallback to other methods)
	os.Unsetenv("CLUSTER_ID")
	clusterID, err = detector.DetectClusterID(context.Background())
	if err != nil {
		t.Errorf("Expected no error with fallback methods, got %v", err)
	}

	if clusterID == "" {
		t.Error("Expected non-empty cluster ID from fallback methods")
	}
	t.Logf("Fallback cluster ID: %s", clusterID)
}

func TestDetectClusterIDNoSources(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create empty clientset with no resources
	clientset := fake.NewSimpleClientset()
	detector := NewClusterIDDetector(clientset, logger)

	_, err := detector.DetectClusterID(context.Background())
	if err == nil {
		t.Error("Expected error when no cluster ID sources available, got nil")
	}
}

func BenchmarkDetectClusterID(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID("12345678-1234-1234-1234-123456789012"),
		},
	}

	clientset := fake.NewSimpleClientset(namespace)
	detector := NewClusterIDDetector(clientset, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectClusterID(context.Background())
	}
}

// TestDetectFromAKS tests Azure AKS cluster detection
func TestDetectFromAKS(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name           string
		node           *corev1.Node
		expectedPrefix string
		expectError    bool
	}{
		{
			name: "Valid AKS node with cluster label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "aks-nodepool1-12345678-vmss000000",
					Labels: map[string]string{
						"kubernetes.azure.com/cluster": "my-aks-cluster",
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: "azure:///subscriptions/sub-id/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/vm-0",
				},
			},
			expectedPrefix: "aks-my-rg-my-aks-cluster",
			expectError:    false,
		},
		{
			name: "AKS node without resource group",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "aks-nodepool1-12345678-vmss000000",
					Labels: map[string]string{
						"kubernetes.azure.com/cluster": "my-aks-cluster",
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: "azure:///vm-0",
				},
			},
			expectedPrefix: "aks-my-aks-cluster",
			expectError:    false,
		},
		{
			name: "Non-AKS node",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "regular-node",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "kind://docker/kind/kind-control-plane",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(tt.node)
			detector := NewClusterIDDetector(clientset, logger)

			clusterID, err := detector.detectFromAKS(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if clusterID != tt.expectedPrefix {
					t.Errorf("Expected cluster ID %s, got %s", tt.expectedPrefix, clusterID)
				}
			}
		})
	}
}