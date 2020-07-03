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
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	guestbooklisters "example.com/foo-controller/generated/webapp/listers/webapp/v1"
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

	// auto kube config discovery:
	// out-of-cluster:
	//  1. env KUBECONFIG
	//  2. flag --kubeconfig
	//  3. ~/.kube/kubeconfig
	// in-cluster:
	//  /var/run/secrets/kubernetes.io/serviceaccount/token
	//  /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
	kubeConfig := ctrl.GetConfigOrDie()

	// clienset
	clientset := guestbookclientset.NewForConfigOrDie(kubeConfig)


	// informers
	informerFactory := guestbookinformers.NewSharedInformerFactory(clientset, time.Minute)
	informers := informerFactory.Webapp().V1().Guestbooks()
	informers.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
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

	lister := informers.Lister()
	fmt.Println("================")
	fmt.Println("before cache sync lister got nothing")
	listExample(lister)

	fmt.Println("================")
	timeout := time.NewTimer(time.Second * 30)
	timeoutCh := make(chan struct{})
	go func() {
		<-timeout.C
		timeoutCh <- struct{}{}
	}()
	if ok := cache.WaitForCacheSync(timeoutCh, informers.Informer().HasSynced); !ok {
		klog.Fatalln("Timeout expired during waiting for caches to sync.")
	}

	fmt.Println("================")
	fmt.Println("after cache sync lister got something")
	guestbooks := listExample(lister)

	fmt.Println("================")
	for _, guestbook := range guestbooks {
		clientsetExample(clientset, guestbook)
	}
	<-stopSignalCh

}

func listExample(lister guestbooklisters.GuestbookLister) []*webappv1.Guestbook {
	guestbooks, err := lister.List(labels.NewSelector())
	if err != nil {
		panic(err)
	}
	fmt.Println("list result:")
	for _, guestbook := range guestbooks {
		fmt.Println(guestbook)
	}
	return guestbooks
}

func clientsetExample(clientset *guestbookclientset.Clientset, guestbook *webappv1.Guestbook) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copy := guestbook.DeepCopy()
	copy.Status.Ok = true
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Guestbook resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := clientset.WebappV1().Guestbooks(copy.Namespace).UpdateStatus(context.TODO(), copy, metav1.UpdateOptions{})
	if err != nil {
		println(err)
	}
}