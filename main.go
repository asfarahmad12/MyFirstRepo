package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Node struct {
	Name string `json:"name"`
}

type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func getNodeList(clientset *kubernetes.Clientset) ([]Node, error) {
	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, len(nodeList.Items))
	for i, n := range nodeList.Items {
		nodes[i] = Node{Name: n.Name}
	}
	return nodes, nil
}

func getPodList(clientset *kubernetes.Clientset) ([]Pod, error) {
	podList, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]Pod, len(podList.Items))
	for i, p := range podList.Items {
		pods[i] = Pod{Name: p.Name, Namespace: p.Namespace}
	}
	return pods, nil
}

func main() {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		nodes, err := getNodeList(clientset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
	})

	http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
		pods, err := getPodList(clientset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pods)
	})

	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
