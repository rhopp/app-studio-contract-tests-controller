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

// Function that creates new ContractTest CR and waits for its succesfull reconciliation (a.k.a. the name appears
// in the status)
//func createNewContractAndWaitForGreet(name string, namespace string, waitTime int) {
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Printf("PANICKING!!!!\n %+v", r)
//		}
//	}()
//	ctx := context.Background()
//	contracttest := &webappv1.ContractTests{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "ContractTest",
//			APIVersion: "webapp.appstudio.qe/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      name,
//			Namespace: namespace,
//		},
//		Spec: webappv1.ContractTestsSpec{
//			ContractName: name,
//			WaitSecs:     waitTime,
//		},
//	}
//	contractTestList := &webappv1.ContractTestsList{}
//	err := k8sClient.List(context.Background(), contractTestList)
//	if err != nil {
//		return
//	}
//
//	for _, c := range contractTestList.Items {
//		fmt.Printf("createNewContractAndWaitForGreet: %#v", c)
//	}
//
//	createError := k8sClient.Create(ctx, contracttest)
//	if createError != nil {
//		fmt.Println(createError)
//		panic(createError)
//	}
//
//	contractTestLookupKey := types.NamespacedName{
//		Namespace: namespace,
//		Name:      name,
//	}
//	fmt.Printf("I'M HERE\n")
//	Eventually(func() (string, error) {
//		fmt.Printf("Waiting for message\n")
//		createdContractTest := &webappv1.ContractTests{}
//		err := k8sClient.Get(ctx, contractTestLookupKey, createdContractTest)
//		if err != nil {
//			return "", err
//		}
//		return createdContractTest.Status.Message, nil
//	}, time.Second*50, time.Second).Should(Equal("Hello " + name))
//}

//var _ = Describe("Verify pact with REACT app", func() {
//	pact := &dsl.Pact{
//		Provider: "MyProvider",
//	}
//
//	It("Verify pact", func() {
//		// Certificate magic - for the mocked service to be able to communicate with kube-apiserver & for authorization
//		caCertPool := x509.NewCertPool()
//		caCertPool.AppendCertsFromPEM(testEnv.Config.CAData)
//		certs, err := tls.X509KeyPair(testEnv.Config.CertData, testEnv.Config.KeyData)
//		if err != nil {
//			panic(err)
//		}
//		tlsConfig := &tls.Config{
//			RootCAs:      caCertPool,
//			Certificates: []tls.Certificate{certs},
//		}
//		// End of certificate magic
//
//		//const pactDir = "/home/rhopp/git/patternfly-react-seed/pact/pacts"
//		rawResults, err := pact.VerifyProviderRaw(pactTypes.VerifyRequest{
//			ProviderBaseURL:            testEnv.Config.Host,
//			ProviderVersion:            "1.0.0",
//			BrokerURL:                  "http://pact-broker-pact-broker.apps.app.studio.adapters-crs-qe.com",
//			BrokerUsername:             "admin",
//			BrokerPassword:             "admin",
//			CustomTLSConfig:            tlsConfig,
//			PublishVerificationResults: true,
//			StateHandlers: pactTypes.StateHandlers{
//				"": func() error {
//					return nil
//				},
//				"contract test is already creted & reconciled": func() error {
//					GinkgoRecover()
//					createNewContractAndWaitForGreet("testname", "default", 5)
//					return nil
//				},
//			},
//			BeforeEach: func() error {
//				err := deleteAllContracts()
//				if err != nil {
//					return err
//				}
//				return nil
//			},
//		})
//		if err != nil {
//			fmt.Printf("Error happened while verifying pacts.\n %#v", err)
//			return
//		}
//		fmt.Println("======== RESULTS =======")
//		for _, result := range rawResults {
//			for _, example := range result.Examples {
//				fmt.Printf("%s: %s\n", example.FullDescription, example.Status)
//			}
//
//		}
//		fmt.Println(rawResults)
//	})
//})
