package controllers

import (
	"errors"
	"fmt"
	"time"

	helmclient "github.com/mabhi/mimic-helm-mvp/helm-client"
	"github.com/mabhi/mimic-helm-mvp/models"
	"github.com/mabhi/mimic-helm-mvp/utility"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	OldKey     = "Old"
	NewKey     = "New"
	DeleteOps  = "Delete"
	UpdateOps  = "Update"
	CreateOps  = "Create"
	maxRetries = 3
)

type HelmController struct {
	//need clientset so we can create resources
	dynamic dynamic.Interface
	//so we can pick up things from it and perform stuff on it
	informer  cache.SharedIndexInformer
	queue     workqueue.RateLimitingInterface
	clientset kubernetes.Interface
	hclient   *helmclient.HelmClient
}

const TargetNamespace = "helm-store"

func NewController(dynamic dynamic.Interface, informer cache.SharedIndexInformer, clientset kubernetes.Interface, hclient *helmclient.HelmClient) *HelmController {
	c := &HelmController{
		dynamic:   dynamic,
		informer:  informer,
		queue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		clientset: clientset,
		hclient:   hclient,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addHandler,
		DeleteFunc: c.deleteHandler,
		// UpdateFunc: c.updateHandler,
	})

	return c
}

func (c *HelmController) Run(ch <-chan struct{}) {
	fmt.Println("Starting the HelmController")
	go c.informer.Run(ch)

	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		fmt.Println("waiting for cache to be synced")
	}

	fmt.Println("helm controller synced and ready")

	wait.Until(c.worker, 2*time.Second, ch)
}

func (c *HelmController) worker() {
	ok, err := c.doProcessing()
	if !ok {
		fmt.Printf("item processed with error %s \n", err.Error())
	}
}

func (hc *HelmController) doProcessing() (bool, error) {
	key, shutdown := hc.queue.Get()
	if shutdown {
		return false, errors.New("queue received shutdown, ending")
	}
	defer hc.queue.Done(key)

	// do your work on the key.
	err := hc.processItem(key.(string))
	if err == nil {
		// No error, tell the queue to stop tracking history
		hc.queue.Forget(key)
	} else if hc.queue.NumRequeues(key) < maxRetries {
		fmt.Printf("Error processing %s (will retry): %v", key, err)
		// requeue the item to work on later
		hc.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		fmt.Printf("Error processing %s (giving up): %v", key, err)
		hc.queue.Forget(key)
	}

	return true, nil
}

func (hc *HelmController) processItem(key string) error {
	fmt.Printf("Processing change to helm resource %s\n", key)
	obj, exists, err := hc.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		hc.handleHelmDeleteAction(obj)
		return nil
	}

	hc.handleHelmInstallAction(obj)
	return nil
}

// handlers
func (c *HelmController) deleteHandler(obj interface{}) {
	fmt.Println("handle delete")
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		fmt.Println("queued delete request")
		c.queue.Add(key)
	} else {
		fmt.Printf("delete handler error %v\n", err.Error())
	}

}

func (c *HelmController) addHandler(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		fmt.Println("queued add request")
		c.queue.Add(key)
	} else {
		fmt.Printf("add handler error %v\n", err.Error())
	}

}

func (c *HelmController) updateHandler(oldObj interface{}, newObj interface{}) {
	fmt.Println("hanle update")

	var updateMap = make(map[string]interface{})
	updateMap[UpdateOps] = map[string]interface{}{
		OldKey: oldObj,
		NewKey: newObj,
	}
	c.queue.Add(updateMap)
}

// processors
func (c *HelmController) handleHelmUpdateAction(oldObj interface{}, newObj interface{}) error {
	crObjWrapper := oldObj.(*unstructured.Unstructured)

	jsonObj := crObjWrapper.Object
	var oldHelmEntity models.HelmAction
	utility.Deserialize(jsonObj, &oldHelmEntity)

	crObjWrapper = newObj.(*unstructured.Unstructured)

	jsonObj = crObjWrapper.Object
	var newHelmEntity models.HelmAction
	utility.Deserialize(jsonObj, &newHelmEntity)

	if oldHelmEntity.Spec.ChartName != newHelmEntity.Spec.ChartName ||
		oldHelmEntity.Spec.RepoUrl != newHelmEntity.Spec.RepoUrl ||
		oldHelmEntity.Spec.RepoName != newHelmEntity.Spec.RepoName {
		return fmt.Errorf("invalid spec updates. Donot change the ChartName, RepoUrl, RepoName. Please provide a newer ChartVersion")

	}

	res, err := c.hclient.InstallApp(newHelmEntity, true)
	if err != nil {
		return fmt.Errorf("upgrade error %v", err.Error())
	} else {
		fmt.Println("Release success ")
		fmt.Printf("Name: %s, Status: %v Release No. %d\n", res.Name, res.Info.Status, res.Version)
	}
	return nil

}

func (c *HelmController) handleHelmDeleteAction(obj interface{}) error {
	crObjWrapper := obj.(*unstructured.Unstructured)

	jsonObj := crObjWrapper.Object
	var helmEntity models.HelmAction

	utility.Deserialize(jsonObj, &helmEntity)
	err := c.hclient.DeleteApp(helmEntity)
	if err != nil {
		return fmt.Errorf("delete error %v", err.Error())
	} else {
		fmt.Println("delete success")
	}
	return nil
}

func (c *HelmController) handleHelmInstallAction(obj interface{}) error {

	crObjWrapper := obj.(*unstructured.Unstructured)

	jsonObj := crObjWrapper.Object
	var helmEntity models.HelmAction

	utility.Deserialize(jsonObj, &helmEntity)
	res, err := c.hclient.InstallApp(helmEntity, false)
	if err != nil {
		return fmt.Errorf("install error %v", err.Error())
	} else {
		fmt.Println("Release success ")
		fmt.Printf("Name: %s, Status: %v Release No. %d\n", res.Name, res.Info.Status, res.Version)
	}
	return nil
}
