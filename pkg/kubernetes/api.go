package kubernetes

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ApiRequest(uri string, method string, body []byte) (statusCode int, data []byte, err error) {
	fmt.Println(uri)
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return 0, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, nil, err
	}

	var res rest.Result

	switch method {
	case "GET":
		fmt.Println("123123")

		res = clientset.RESTClient().Get().RequestURI(uri).Do()
	case "POST":
		res = clientset.RESTClient().Get().Body(body).RequestURI(uri).Do()
	case "PUT":
		res = clientset.RESTClient().Get().Body(body).RequestURI(uri).Do()
	case "DELETE":
		res = clientset.RESTClient().Get().Body(body).RequestURI(uri).Do()
	default:
		err = errors.New("unsupported REST method")
	}

	data, err = res.Raw()
	if err != nil {
		return 0, nil, err
	}
	res.StatusCode(&statusCode)

	return statusCode, data, nil
}
