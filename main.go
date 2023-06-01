package main

import (
	"flag"
	"fmt"
	"time"

	ct "github.com/mabhi/mimic-helm-mvp/controllers"
	helmclient "github.com/mabhi/mimic-helm-mvp/helm-client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config := flag.String("kubeconfig", "/home/abhijeet/.kube/config", "this is where the kubeconfig file is stored")
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", *config)
	if err != nil {
		fmt.Printf("Creating kubeconfig file: %s", err.Error())

		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Creating in cluster config:%s", err.Error())
		}
	}

	//dynamic clientset
	dynclient, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("creating dynamic client error : %s", err.Error())
	}
	//typed client set
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Printf("creating client set error : %s", err.Error())
	}
	hc := helmclient.NewClient(kubeconfig, ct.TargetNamespace)

	resource := schema.GroupVersionResource{
		Group:    "mabhi.dev",
		Version:  "v1",
		Resource: "helm-actions",
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynclient, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()

	controller := ct.NewController(dynclient, informer, clientset, hc)

	// For reconciliation we need the following setup. Skipping for now
	/*
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
		if err != nil {
			fmt.Println(err, "could not create manager")
			os.Exit(1)
		}
		ctrl.NewControllerManagedBy(mgr).
			Named("Controller").
			For(&v1.Pod{}).
			Complete(&ct.Reconciler{})
	*/
	fmt.Println("controller started")

	ch := make(chan struct{})
	controller.Run(ch)
	defer close(ch)

}
