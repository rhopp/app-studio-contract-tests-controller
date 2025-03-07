/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "appstudio.qe/contract-tests/api/v1"
)

// ContractTestsReconciler reconciles a ContractTests object
type ContractTestsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.appstudio.qe,resources=contracttests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.appstudio.qe,resources=contracttests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.appstudio.qe,resources=contracttests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ContractTests object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ContractTestsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	// var helloWorld examplev1.HelloWorld
	var ContractT webappv1.ContractTests
	if err := r.Get(ctx, req.NamespacedName, &ContractT); err != nil {
		log.Error(err, "unable to fetch Contract")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Reconciler Contract name: " + ContractT.Spec.ContractName)

	log.Info("Waiting for secs:" + fmt.Sprint(ContractT.Spec.WaitSecs))
	// var mostRecentTime *time.Time

	time.Sleep(time.Duration(ContractT.Spec.WaitSecs) * time.Second)
	// var NowTime *v1.Time

	ContractT.Status.Message = "Hello " + ContractT.Spec.ContractName
	// ContractT.Status.UpdatedAt = nil

	ContractT.Status.UpdatedAt = time.Now().String()
	ContractT.Status.Status = "Active"

	if err := r.Status().Update(ctx, &ContractT); err != nil {
		log.Error(err, "Unable to update HelloWorld status")
		return ctrl.Result{}, err
	}
	log.Info("Status updated")
	log.Info(fmt.Sprintf("%+v\n", ContractT))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContractTestsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.ContractTests{}).
		Complete(r)
}
