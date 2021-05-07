/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// A small utility program to lookup hostnames of endpoints in a service.
package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	//
)

var (
	onStart    = flag.String("on-start", "", "Script to run on start, must accept a new line separated list of peers via stdin.")
	namespace  = flag.String("ns", "", "The namespace this pod is running in. If unspecified, the POD_NAMESPACE env var is used.")
	labelSelector = flag.String("labels", "", "Label selector")
)

func shellOut(sendStdin, script string) {
	log.Printf("execing: %v with stdin: %v", script, sendStdin)
	// TODO: Switch to sending stdin from go
	out, err := exec.Command("bash", "-c", fmt.Sprintf("echo -e '%v' | %v", sendStdin, script)).CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to execute %v: %v, err: %v", script, string(out), err)
	}
	log.Print(string(out))
}

func extractIps(list []string) []string {
	result := make([]string, len(list))
	for i, _ := range list {
		parts := strings.Split(list[i], ".")
		result[i] = strings.Replace(parts[0], "-", ".", 3)
	}

	return result
}

func main() {
	flag.Parse()

	ns := *namespace
	if ns == "" {
		ns = os.Getenv("NAMESPACE")
	}
	script := *onStart

	if ns == "" && labelSelector == nil ||  *onStart == "" {
		log.Fatalf("Incomplete args, require -on-change and/or -on-start, -service and -ns or an env var for POD_NAMESPACE.")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	log.Printf("Using ns %s labels %s", ns, *labelSelector)
	peerList := make([]string, 0)
	emptyIp := true
	for emptyIp  {
		pods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{
			LabelSelector: *labelSelector,
		})
		log.Printf("PODs len %d \n", len(pods.Items))
		if err != nil {
			panic(err.Error())
		}
		counterIps := 0
		for _, pod := range pods.Items {
			if pod.Status.PodIP == "" {
				log.Printf("POD IP is empty\n")
			} else {
				peerList = append(peerList, pod.Status.PodIP)
				counterIps = counterIps + 1
			}
		}
		if counterIps == len(pods.Items) {
			emptyIp = false
		}
		time.Sleep(10*time.Second)
	}

	if len(peerList) == 0 {
		log.Printf("List is empty")
		panic("list should not be empty")
	}

	sort.Strings(peerList)
	log.Printf("Peer list updated was %v len %d \n", peerList, len(peerList))
	shellOut(strings.Join(peerList, "\n"), script)
	log.Printf("Peer finder exiting")
}
