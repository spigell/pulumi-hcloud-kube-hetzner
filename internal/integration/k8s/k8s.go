package k8s

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type K8S struct {
	ctx context.Context

	Client *kubernetes.Clientset
}

func New(ctx context.Context, kubeconfig string) (*K8S, error) {
	client, err := NewClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &K8S{
		ctx:    ctx,
		Client: client,
	}, nil
}
