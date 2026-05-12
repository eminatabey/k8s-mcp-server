package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset *kubernetes.Clientset
	Mode      string
}

func NewClient() (*Client, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("in-cluster clientset: %w", err)
		}
		return &Client{Clientset: cs, Mode: "in-cluster"}, nil
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig %q: %w", kubeconfig, err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig clientset: %w", err)
	}
	return &Client{Clientset: cs, Mode: "kubeconfig"}, nil
}
