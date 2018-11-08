package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ApiRequest(uri string, method string, body string) (result string, err error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	byteArray, err := clientset.RESTClient().Get().RequestURI(uri).Do().Raw()
	if err != nil {
		return "", err
	}

	result = string(byteArray[:])

	fmt.Println(result)

	return result, nil

	//restClient := clientset.CoreV1().RESTClient()

	//restClient.Get()

	//for {
	//	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	//
	//	// Examples for error handling:
	//	// - Use helper functions like e.g. errors.IsNotFound()
	//	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	//	_, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
	//	if errors.IsNotFound(err) {
	//		fmt.Printf("Pod not found\n")
	//	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	//		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	//	} else if err != nil {
	//		panic(err.Error())
	//	} else {
	//		fmt.Printf("Found pod\n")
	//	}
	//
	//	time.Sleep(10 * time.Second)
	//}

}
