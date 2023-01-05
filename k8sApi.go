package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func getKubeClients(fileData []byte) (*kubernetes.Clientset, *metrics.Clientset, error) {
	clientCfg, _ := clientcmd.NewClientConfigFromBytes(fileData)
	cfg, _ := clientCfg.ClientConfig()

	kubeClient, _ := kubernetes.NewForConfig(cfg)
	metricsClient, _ := metrics.NewForConfig(cfg)
	return kubeClient, metricsClient, nil
}

func pods(kc *kubernetes.Clientset, k8op map[string][]string) error {
	pods_list, err := kc.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("error getting pods: " + err.Error())
		return err
	}
	for _, pod := range pods_list.Items {

		podstatusPhase := string(pod.Status.Phase)
		podCreationTime := pod.GetCreationTimestamp()
		age := time.Since(podCreationTime.Time).Round(time.Second)
		temp := []string{
			pod.GetName(),
			pod.GetNamespace(),
			pod.OwnerReferences[0].Kind,
			podstatusPhase,
			fmt.Sprintf("%d", restartCount(&pod)),
			age.String(),
			podCreationTime.String(),
		}
		key := pod.GetName() + "_" + pod.GetNamespace()
		k8op[key] = temp
	}
	return nil

}

func podMetrics(mc *metrics.Clientset, k8op map[string][]string) error {
	podMetricList, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("error getting pod metrics: " + err.Error())
		return err
	}
	for _, podMetrics := range podMetricList.Items {

		podName := podMetrics.ObjectMeta.GetName()
		key := podName + "_" + podMetrics.GetNamespace()

		podContainers := podMetrics.Containers
		var cont_total_cpu = &inf.Dec{}
		var cont_total_mry = &inf.Dec{}

		for _, container := range podContainers {
			cpu := container.Usage.Cpu().AsDec()
			mry := container.Usage.Memory().AsDec()

			cont_total_cpu.Add(cont_total_cpu, cpu)
			cont_total_mry.Add(cont_total_mry, mry)

		}

		k8op[key] = append(k8op[key], fmt.Sprintf("%v", cont_total_cpu.String()))
		k8op[key] = append(k8op[key], fmt.Sprintf("%v", cont_total_mry.String()))
		//k8op[key]["memory"] = fmt.Sprintf("%v", cont_total_mry.String())
	}
	return nil
}

func restartCount(pod *corev1.Pod) int32 {
	if len(pod.Status.ContainerStatuses) > 0 {
		return pod.Status.ContainerStatuses[0].RestartCount
	}
	return 0
}
