package controllers

import (
	webappv1 "appstudio.qe/contract-tests/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("Contract Tests Controller", func() {
	const (
		ctName      = "my-name"
		ctNamespace = "default"
	)
	It("Should update status after timeout", func() {
		contracttest := &webappv1.ContractTests{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ContractTest",
				APIVersion: "webapp.appstudio.qe/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      ctName,
				Namespace: ctNamespace,
			},
			Spec: webappv1.ContractTestsSpec{
				ContractName: ctName,
				WaitSecs:     5,
			},
		}
		Expect(k8sClient.Create(ctx, contracttest)).Should(Succeed())

		contracttestLookupKey := types.NamespacedName{Namespace: "default", Name: ctName}
		contracttestCR := &webappv1.ContractTests{}
		startTime := time.Now()
		Eventually(func() (string, error) {
			err := k8sClient.Get(ctx, contracttestLookupKey, contracttestCR)
			if err != nil {
				return "", err
			}
			return contracttestCR.Status.Message, nil
		}, time.Second*30, time.Second).Should(Equal("Hello " + ctName))
		endTime := time.Now()
		Expect(endTime).ShouldNot(BeTemporally("~", startTime, 5*time.Second))
	})
})
