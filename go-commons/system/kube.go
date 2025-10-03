package system

import (
	"context"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	Client *kubernetes.Clientset
}

func NewKubeClient() (*kubernetes.Clientset, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return kubernetes.NewForConfig(cfg)
	}
	kcfg := os.Getenv("KUBECONFIG")
	if kcfg == "" {
		home, _ := os.UserHomeDir()
		kcfg = home + "/.kube/config"
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kcfg)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func (k *Kube) ListPods(ctx context.Context, ns, ls string) ([]v1.Pod, error) {
	pods, err := k.Client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: ls})
	if err != nil {
		return nil, err
	}
	var out []v1.Pod
	for _, p := range pods.Items {
		if p.Status.PodIP != "" && p.Status.Phase == v1.PodRunning {
			out = append(out, p)
		}
	}
	return out, nil
}
