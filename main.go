/*


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

package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"time"

	webappv1 "example.com/foo-controller/apis/webapp/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports

	guestbookclientset "example.com/foo-controller/generated/webapp/clientset/versioned"
	guestbookinformers "example.com/foo-controller/generated/webapp/informers/externalversions"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = webappv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// apply CRD first:
//  kubectl apply -f config/crd/bases/webapp.example.com_guestbooks.yaml
//  kubectl apply -f config/samples/webapp_v1_guestbook.yaml
// then run this program
func main() {

	// stop signal channel which is triggered for SIGTERM or SIGINT
	stopSignalCh := ctrl.SetupSignalHandler()

	// get config from:
	// out-of-cluster:
	//  1. env KUBECONFIG
	//  2. flag --kubeconfig
	// in-cluster:
	//  /var/run/secrets/kubernetes.io/serviceaccount/token
	//  /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
	kubeConfig := ctrl.GetConfigOrDie()

	// clienset
	clientset := guestbookclientset.NewForConfigOrDie(kubeConfig)

	// informers
	informerFactory := guestbookinformers.NewSharedInformerFactory(clientset, time.Minute)
	informer := informerFactory.Webapp().V1().Guestbooks()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			klog.Infof("Added: %v", object)
		},
		UpdateFunc: func(oldObject, newObject interface{}) {
			klog.Infof("Updated: %v", newObject)
		},
		DeleteFunc: func(object interface{}) {
			klog.Infof("Deleted: %v", object)
		},
	})

	informerFactory.Start(stopSignalCh)

	lister := informer.Lister()

	timeout := time.NewTimer(time.Second * 30)
	timeoutCh := make(chan struct{})
	go func() {
		<-timeout.C
		timeoutCh <- struct{}{}
	}()
	if ok := cache.WaitForCacheSync(timeoutCh, informer.Informer().HasSynced); !ok {
		klog.Fatalln("Timeout expired during waiting for caches to sync.")
	}

	guestbooks, err := lister.List(labels.NewSelector())
	if err != nil {
		panic(err)
	}
	for _, guestbook := range guestbooks {
		fmt.Println(guestbook)
	}

}
