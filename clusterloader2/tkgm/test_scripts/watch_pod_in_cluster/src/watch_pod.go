package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type ObjectSelector struct {
	Namespace     string
	LabelSelector string
	FieldSelector string
}

func NewInformer(
	c clientset.Interface,
	kind string,
	selector *ObjectSelector,
	handleObj func(interface{}, interface{}),
) cache.SharedInformer {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = selector.FieldSelector
		options.LabelSelector = selector.LabelSelector
	}
	listerWatcher := cache.NewFilteredListWatchFromClient(c.CoreV1().RESTClient(), kind, selector.Namespace, optionsModifier)
	informer := cache.NewSharedInformer(listerWatcher, nil, 0)
	addEventHandler(informer, handleObj)

	return informer
}

func addEventHandler(i cache.SharedInformer,
	handleObj func(interface{}, interface{}),
) {
	i.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleObj(nil, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			handleObj(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				handleObj(tombstone.Obj, nil)
			} else {
				handleObj(obj, nil)
			}
		},
	})
}

// StartAndSync starts informer and waits for it to be synced.
func StartAndSync(i cache.SharedInformer, stopCh chan struct{}, timeout time.Duration) error {
	go i.Run(stopCh)
	timeoutCh := make(chan struct{})
	timeoutTimer := time.AfterFunc(timeout, func() {
		close(timeoutCh)
	})
	defer timeoutTimer.Stop()
	if !cache.WaitForCacheSync(timeoutCh, i.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}
func createMetaNamespaceKey(namespace, name string) string {
	return namespace + "/" + name
}

func checkPod(_, obj interface{}) {

	if obj == nil {
		return
	}
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	if pod.Status.Phase == corev1.PodRunning {
		key := createMetaNamespaceKey(pod.Namespace, pod.Name)

		if _, found := podStartupEntries[key]; !found {
			podStartupEntries[key] = make(map[string]time.Time)
			podStartupEntries[key][watchPhase] = time.Now()

			podStartupEntries[key][createPhase] = pod.CreationTimestamp.Time
			diff := podStartupEntries[key][watchPhase].Sub(podStartupEntries[key][createPhase])
			log.Println(createPhase, "---------", key, "--------", podStartupEntries[key][createPhase])
			log.Println(watchPhase, " ---------", key, "--------", podStartupEntries[key][watchPhase])
			log.Println("watch_to_create", "---", key, "--------", diff)

			/*			var startTime metav1.Time
						for _, cs := range pod.Status.ContainerStatuses {
							if cs.State.Running != nil {
								if startTime.Before(&cs.State.Running.StartedAt) {
									startTime = cs.State.Running.StartedAt
								}
							}
						}
						if startTime != metav1.NewTime(time.Time{}) {
							p.podStartupEntries.Set(key, runPhase, startTime.Time)
						} else {
							klog.Errorf("%s: pod %v (%v) is reported to be running, but none of its containers is", p, pod.Name, pod.Namespace)
						}*/
		}
	}
}

const (
	//defaultPodStartupLatencyThreshold = 5 * time.Second
	informerSyncTimeout = time.Minute

	createPhase = "create"
	//schedulePhase = "schedule"
	//runPhase      = "run"
	watchPhase = "watch"
)

var podStartupEntries = make(map[string]map[string]time.Time)

func main() {
	log.Println("Hello, world.")
	stopCh := make(chan struct{})
	selector := &ObjectSelector{
		Namespace:     metav1.NamespaceAll,
		LabelSelector: "group = latency",
		FieldSelector: "",
	}
	/*var kubeconfig = "/root/.kube/config"


	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
		// create the clientset
	c, err := kubernetes.NewForConfig(config) */

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	go func() {
		select {
		case sig := <-ch:
			log.Println("Got %s signal. Aborting...\n", sig)
			for pod, _ := range podStartupEntries {
				create := podStartupEntries[pod][createPhase]
				watch := podStartupEntries[pod][watchPhase]
				diff := watch.Sub(create)
				log.Println("-------------")
				log.Println(pod, createPhase, create)
				log.Println(pod, watchPhase, watch)
				log.Println(pod, "watch_to_create", diff)
			}
			os.Exit(1)
		}
	}()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	i := NewInformer(
		c,
		"pods",
		selector,
		checkPod,
	)
	for {
		StartAndSync(i, stopCh, informerSyncTimeout)
	}

}
