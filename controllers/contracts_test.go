package controllers

import (
	webappv1 "appstudio.qe/contract-tests/api/v1"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/pact-foundation/pact-go/dsl"
	pactTypes "github.com/pact-foundation/pact-go/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
)

var (
	g       *gomega.WithT
	cancel1 context.CancelFunc
)

func TestContracts(t *testing.T) {
	g = gomega.NewGomegaWithT(t)

	err := setupTestEnv(t)
	if err != nil {
		t.Errorf("Failed to setup TestEnv. \n%+v", err)
	}

	pact := &dsl.Pact{
		Provider: "MyProvider",
	}

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

	_, err = pact.VerifyProvider(t, pactTypes.VerifyRequest{
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
		t.Errorf("Error while verifying tests. \n %+v", err)
	}

	cancel1()
	err = testEnv.Stop()
	if err != nil {
		fmt.Println("Stopping failed")
		fmt.Printf("%+v", err)
		panic("Cleanup failed")
	}

}

func setupTestEnv(t *testing.T) error {

	ctx, cancel1 = context.WithCancel(context.TODO())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		return err
	}

	err = webappv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}

	err = (&ContractTestsReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			cancel1()
			fmt.Println("Stopping testEnv")
			err := testEnv.Stop()
			if err != nil {
				panic("Cleanup failed")
			}
		}()
		err = k8sManager.Start(ctx)
		if err != nil {
			t.Errorf("Error starting k8Manager")
		}
	}()
	return nil
}

func createNewContractAndWaitForGreet(name string, namespace string, waitTime int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANICKING!!!!\n %+v", r)
		}
	}()
	//ctx, cancel := context.WithCancel(context.TODO())
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
	err := k8sClient.List(ctx, contractTestList)
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
	g.Eventually(func() (string, error) {
		fmt.Printf("Waiting for message\n")
		createdContractTest := &webappv1.ContractTests{}
		err := k8sClient.Get(ctx, contractTestLookupKey, createdContractTest)
		if err != nil {
			return "", err
		}
		return createdContractTest.Status.Message, nil
	}, time.Second*50, time.Second).Should(gomega.Equal("Hello " + name))
}

// Helper function to delete all contracts. This is called in BeforeAll contracts to avoid conflicts.
func deleteAllContracts() error {
	fmt.Println("Deleting all contracts")

	contractsList := &webappv1.ContractTestsList{}
	err := k8sClient.List(ctx, contractsList)
	if err != nil {
		return err
	}
	if len(contractsList.Items) > 0 {
		fmt.Println("ContractList.Items > 0, Items:")
		for _, c := range contractsList.Items {

			fmt.Printf("Deleting: %#v", c)
			err := k8sClient.Delete(ctx, &c)
			if err != nil {
				fmt.Printf("Error deleting: %#v", err)
				return err
			}
		}
	}

	err = wait.PollImmediate(10*time.Second, 250*time.Millisecond, func() (done bool, err error) {
		internalContractList := &webappv1.ContractTestsList{}
		fmt.Println("Waiting for contracts to be deleted")
		err = k8sClient.List(ctx, internalContractList)
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
