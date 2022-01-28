package controllers

import (
	webappv1 "appstudio.qe/contract-tests/api/v1"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pact-foundation/pact-go/dsl"
	pactTypes "github.com/pact-foundation/pact-go/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

// Helper function to delete all contracts. This is called in BeforeAll contracts to avoid conflicts.
func deleteAllContracts() error {
	fmt.Println("Deleting all contracts")

	contractsList := &webappv1.ContractTestsList{}
	err := k8sClient.List(context.Background(), contractsList)
	if err != nil {
		return err
	}
	if len(contractsList.Items) > 0 {
		fmt.Println("ContractList.Items > 0, Items:")
		for _, c := range contractsList.Items {

			fmt.Printf("Deleting: %#v", c)
			err := k8sClient.Delete(context.Background(), &c)
			if err != nil {
				fmt.Printf("Error deleting: %#v", err)
				return err
			}
		}
	}

	err = wait.PollImmediate(10*time.Second, 250*time.Millisecond, func() (done bool, err error) {
		internalContractList := &webappv1.ContractTestsList{}
		fmt.Println("Waiting for contracts to be deleted")
		err = k8sClient.List(context.Background(), internalContractList)
		if err != nil {
			fmt.Printf("ERROR in PollImmediate")
			return false, err
		}
		if len(internalContractList.Items) > 0 {
			fmt.Println("Remaining contracts:")
			for _, c := range internalContractList.Items {
				fmt.Printf("%#v", c)
			}
			return false, nil
		} else {
			fmt.Println("Contracts deleted!")
			return true, nil
		}
	})
	if err != nil {
		return err
	}
	return nil
}

// Function that creates new ContractTest CR and waits for its succesfull reconciliation (a.k.a. the name appears
// in the status)
func createNewContractAndWaitForGreet(name string, namespace string, waitTime int) {
	defer func() {
		if recover() != nil {
			fmt.Printf("PANICKING!!!!")
		}
	}()
	ctx := context.Background()
	contracttest := &webappv1.ContractTests{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContractTest",
			APIVersion: "webapp.appstudio.qe/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: webappv1.ContractTestsSpec{
			ContractName: name,
			WaitSecs:     waitTime,
		},
	}
	contractTestList := &webappv1.ContractTestsList{}
	err := k8sClient.List(context.Background(), contractTestList)
	if err != nil {
		return
	}

	for _, c := range contractTestList.Items {
		fmt.Printf("createNewContractAndWaitForGreet: %#v", c)
	}

	createError := k8sClient.Create(ctx, contracttest)
	if createError != nil {
		fmt.Println(createError)
		panic(createError)
	}

	contractTestLookupKey := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	Eventually(func() (string, error) {
		createdContractTest := &webappv1.ContractTests{}
		err := k8sClient.Get(ctx, contractTestLookupKey, createdContractTest)
		if err != nil {
			return "", err
		}
		return createdContractTest.Status.Message, nil
	}, time.Second*30, time.Second).Should(Equal("Hello " + name))
}

var _ = Describe("Verify pact with REACT app", func() {
	pact := &dsl.Pact{
		Provider: "MyProvider",
	}

	It("Verify pact", func() {
		// Certificate magic - for the mocked service to be able to communicate with kube-apiserver & for authorization
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(testEnv.Config.CAData)
		certs, err := tls.X509KeyPair(testEnv.Config.CertData, testEnv.Config.KeyData)
		if err != nil {
			panic(err)
		}
		tlsConfig := &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{certs},
		}
		// End of certificate magic

		//const pactDir = "/home/rhopp/git/patternfly-react-seed/pact/pacts"
		rawResults, err := pact.VerifyProviderRaw(pactTypes.VerifyRequest{
			ProviderBaseURL:            testEnv.Config.Host,
			ProviderVersion:            "1.0.0",
			BrokerURL:                  "http://pact-broker-pact-broker.apps.app.studio.adapters-crs-qe.com",
			BrokerUsername:             "admin",
			BrokerPassword:             "admin",
			CustomTLSConfig:            tlsConfig,
			PublishVerificationResults: true,
			StateHandlers: pactTypes.StateHandlers{
				"": func() error {
					return nil
				},
				"contract test is already creted & reconciled": func() error {
					GinkgoRecover()
					createNewContractAndWaitForGreet("testname", "default", 5)
					return nil
				},
			},
			BeforeEach: func() error {
				err := deleteAllContracts()
				if err != nil {
					return err
				}
				return nil
			},
		})
		if err != nil {
			fmt.Printf("Error happened while verifying pacts.\n %#v", err)
			return
		}
		fmt.Println("======== RESULTS =======")
		for _, result := range rawResults {
			for _, example := range result.Examples {
				fmt.Printf("%s: %s\n", example.FullDescription, example.Status)
			}

		}
		fmt.Println(rawResults)
	})
})
