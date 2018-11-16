package kubernetes

import (
	"github.com/pkg/errors"
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

	var byteArray []byte

	switch method {
	case "GET":
		byteArray, err = clientset.RESTClient().Get().RequestURI(uri).Do().Raw()
	case "POST":
		byteArray, err = clientset.RESTClient().Post().RequestURI(uri).Body(body).Do().Raw()
	case "PUT":
		byteArray, err = clientset.RESTClient().Put().RequestURI(uri).Body(body).Do().Raw()
	case "DELETE":
		byteArray, err = clientset.RESTClient().Delete().RequestURI(uri).Body(body).Do().Raw()
	default:
		err = errors.New("unsupported REST method")
	}

	if err != nil {
		return "", err
	}

	result = string(byteArray[:])

	return result, nil
}
