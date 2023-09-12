package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	refresh_period = 10 //10sec
)

type KubeConfig struct {
	//generate online (use yml -> go struct)
}

func importConfig(c *gin.Context) {

	mf, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "please upload the file")
		return
	}
	file, err := mf.Open()
	if err != nil {
		c.String(http.StatusBadRequest, "cannot open the file")
		return
	}

	byt, err := ioutil.ReadAll(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "error reading the file")
		return
	}

	kubeClient, metricsClient, _ := getKubeClients(byt)
	ws, _ := NewWebSocket(c.Writer, c.Request)
	fetchDataForWs := func(arguments ...any) (any, error) {
		kubeClient := arguments[0].(*kubernetes.Clientset)
		metricsClient := arguments[1].(*metrics.Clientset)
		data := map[string][]string{}
		err := pods(kubeClient, data)
		if err != nil {
			return nil, err
		}
		err = podMetrics(metricsClient, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	go writeDatatoWS(ws, fetchDataForWs, refresh_period, kubeClient, metricsClient)

}
