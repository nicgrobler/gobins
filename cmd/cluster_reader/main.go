package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	coreTypes "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	rbac "k8s.io/client-go/kubernetes/typed/rbac/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type resultList struct {
	RoleBindings []string                 `json:"rolebindings"`
	Quotas       []coreTypes.ResourceList `json:"quotas"`
}

type namespaceList struct {
	NameSpaces []string `json:"namespaces"`
}

type bindindingsList struct {
	RoleBindings []string `json:"rolebindings"`
}

type quotasList struct {
	Quotas []coreTypes.ResourceList `json:"quotas"`
}

func getNamespaceList(client *core.CoreV1Client) (namespaceList, error) {
	list := namespaceList{}
	spaces, err := client.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return list, err
	}
	for _, space := range spaces.Items {
		list.NameSpaces = append(list.NameSpaces, space.Name)
	}
	return list, nil
}

func getRoleBindings(client *rbac.RbacV1Client, namespace string) (bindindingsList, error) {

	list := bindindingsList{}
	bindings, err := client.RoleBindings(namespace).List(metav1.ListOptions{})
	if err != nil {
		return list, err
	}
	for _, item := range bindings.Items {
		for _, subject := range item.Subjects {
			list.RoleBindings = append(list.RoleBindings, subject.Name)
		}

	}

	return list, nil
}

func getQuotas(client *core.CoreV1Client, namespace string) (quotasList, error) {

	list := quotasList{}
	quotas, err := client.ResourceQuotas(namespace).List(metav1.ListOptions{})
	if err != nil {
		return list, err
	}
	for _, item := range quotas.Items {
		list.Quotas = append(list.Quotas, item.Spec.Hard)
	}

	return list, nil
}

func isNamespacePresent(nameSpace string, nameSpaces []string) bool {
	for _, name := range nameSpaces {
		if name == nameSpace {
			return true
		}
	}
	return false
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func getConfig(kubeconfig, clusterContext, nameSpace string, localOnly bool) *rest.Config {
	var config *rest.Config
	if localOnly {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		return config
	}
	if kubeconfig == "" || clusterContext == "" || nameSpace == "" {
		log.Fatalf("kubeconfig: %s, context: %s, and namespace: %s\n", kubeconfig, clusterContext, nameSpace)
	}

	// use the current context in kubeconfig
	config, err := buildConfigFromFlags(clusterContext, kubeconfig)
	if err != nil {
		log.Fatalf("buidling config failed: %s", err.Error())
	}
	return config
}

func main() {
	var kubeconfig *string
	var clusterContext *string
	var nameSpace *string
	var localOnly *bool

	localOnly = flag.Bool("local", false, "can bypass kubeconfig requirement if running within pod that has service account.")
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	clusterContext = flag.String("context", "", "specify the context to use for the connection")
	nameSpace = flag.String("namespace", "", "the namespace to query against")

	flag.Parse()

	// get config corresponding to chosen flags
	config := getConfig(*kubeconfig, *clusterContext, *nameSpace, *localOnly)
	results := resultList{}

	// create the clientsets
	clientset, err := core.NewForConfig(config)
	if err != nil {
		log.Fatalf("generate client from config failed: %s", err.Error())
	}
	clientRBAC, err := rbac.NewForConfig(config)
	if err != nil {
		log.Fatalf("generate rbac client from config failed: %s", err.Error())
	}

	// if the namespace is not present, find out now, and bail if not
	namespaces, err := getNamespaceList(clientset)
	if err != nil {
		log.Fatalf("generate list of namespaces failed: %s", err.Error())
	}

	if !isNamespacePresent(*nameSpace, namespaces.NameSpaces) {
		log.Fatal("supplied namespace not found\n")
	}

	// now grab the roleBinndings and resourceQuotas from this namespace
	bindings, err := getRoleBindings(clientRBAC, *nameSpace)
	if err != nil {
		log.Fatalf("generate list of role bindings failed: %s", err.Error())
	}

	results.RoleBindings = bindings.RoleBindings

	quotas, err := getQuotas(clientset, *nameSpace)
	if err != nil {
		log.Fatalf("generate list of role bindings failed: %s", err.Error())
	}

	results.Quotas = quotas.Quotas

	resultString, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("failed to encode json due to error: %s", err.Error())
	}

	fmt.Println(string(resultString))
}
