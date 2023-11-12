package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func KubeletVersion(kubeconfig string, node string) (string, error) {
	cli, err := Client(kubeconfig)

	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()
	n, err := cli.CoreV1().Nodes().Get(ctx, node, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return n.Status.NodeInfo.KubeletVersion, nil
}
