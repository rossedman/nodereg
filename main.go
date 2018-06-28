package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client     *kubernetes.Clientset
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchList := cache.NewListWatchFromClient(client.Core().RESTClient(), "nodes", metav1.NamespaceAll, fields.Everything())

	informer := cache.NewSharedIndexInformer(
		watchList,
		&api.Node{},
		time.Second*10,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: handleNodeAdd,
		UpdateFunc: func(old, new interface{}) {
			// check if its the same resource version before processing
			newNode := new.(*api.Node)
			oldNode := old.(*api.Node)
			if newNode.ResourceVersion == oldNode.ResourceVersion {
				return
			}

			handleNodeAdd(new)
		},
	})

	stop := make(chan struct{})
	defer close(stop)

	informer.Run(stop)

	// run forever...
	select {}
}

func handleNodeAdd(obj interface{}) {
	node := obj.(*api.Node)
	if !nodeHasRegistered(node.Annotations) {
		if nodeNeedsToRegister(node.Annotations) {
			nodeBytes, _ := json.Marshal(&node)
			nodeToSend := bytes.NewReader(nodeBytes)

			logrus.Infof("registering node %v", node.Name)

			resp, err := http.Post(node.Annotations["rossedman.io/register"], "application/json", nodeToSend)
			if err != nil {
				logrus.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				// update annotation to say that its registered
				node.Annotations["rossedman.io/registered"] = "true"
				_, err := client.CoreV1().Nodes().Update(node)
				if err != nil {
					logrus.Fatal(err)
				}
			}
		}
	} else {
		logrus.Infof("node has registered: %v", node.Name)
	}
}

func nodeHasRegistered(annotations map[string]string) bool {
	for k := range annotations {
		if k == "rossedman.io/registered" {
			return true
		}
	}

	return false
}

func nodeNeedsToRegister(annotations map[string]string) bool {

	for k := range annotations {
		if k == "rossedman.io/register" {
			return true
		}
	}

	return false
}
