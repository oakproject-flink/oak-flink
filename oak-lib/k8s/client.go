package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient creates a Kubernetes client
// It tries in-cluster config first, then falls back to kubeconfig
func NewClient() (kubernetes.Interface, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetConfig returns a Kubernetes REST config
// Tries in-cluster config first, then kubeconfig
func GetConfig() (*rest.Config, error) {
	// Try in-cluster config first (for running inside K8s)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig
	return GetKubeconfigConfig()
}

// GetKubeconfigConfig loads config from kubeconfig file
func GetKubeconfigConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// IsInCluster checks if the code is running inside a Kubernetes cluster
func IsInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}
