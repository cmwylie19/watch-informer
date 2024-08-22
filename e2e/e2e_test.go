//go:build e2e
// +build e2e

package e2e_test

import (
	"bytes"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var kubeConfigPath string

var _ = BeforeSuite(func() {
	kubeConfigPath = setupKindCluster()
	buildImage()
	importImage()
	deployApplication(kubeConfigPath)
	deployWatchedPod(kubeConfigPath)
	buildCurlPod()
	deployCurlPod(kubeConfigPath)
})

// var _ = AfterSuite(func() {
// 	teardownKindCluster(kubeConfigPath)
// })

var _ = Describe("E2E Test", func() {
	Context("When deploying the application", func() {
		It("should deploy successfully and produce logs", func() {

			cmd := exec.Command("kubectl", "exec", "it", "curler", "-n watch-informer", "--", "grpcurl", "-plaintext", "-d", "'{\"group\": \"\", \"version\": \"v1\", \"resource\": \"pod\", \"namespace\": \"default\"}'", "watch-informer.watch-informer.svc.cluster.local:50051", "api.WatchService.Watch")
			var cmdOut bytes.Buffer
			cmd.Stdout = &cmdOut
			cmd.Stderr = &cmdOut
			err := cmd.Run()
			if err != nil {
				GinkgoWriter.Write(cmdOut.Bytes())
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(cmdOut.String()).To(ContainSubstring("EventType: ADD"))
			Expect(cmdOut.String()).To(ContainSubstring("kind:Pod"))

			podLogs := getPodLogs(kubeConfigPath)
			Expect(podLogs).To(ContainSubstring("Server listening at :50051"))
			Expect(podLogs).To(ContainSubstring("Starting watch for -v1-pods-default"))
			Expect(podLogs).To(ContainSubstring("EventType: ADD"))
			Expect(podLogs).To(ContainSubstring("kind:Pod"))
		})
	})
})

func deployCurlPod(kubeConfigPath string) {
	cmd := exec.Command("kubectl", "apply", "-f", "../hack/curler.yaml", "--kubeconfig", kubeConfigPath)
	var cmdOut bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdOut
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(cmdOut.Bytes())
	}
}
func importCurlPod() {
	cmd := exec.Command("kind", "load", "docker-image", "curler:ci")
	var cmdOut bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdOut
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(cmdOut.Bytes())
	}
}
func buildCurlPod() {
	cmd := exec.Command("docker", "build", "-t", "curler:ci", "-f", "--hack/Dockerfile", ".")
	var cmdOut bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdOut
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(cmdOut.Bytes())
	}
}
func deployWatchedPod(kubeConfigPath string) {
	cmd := exec.Command("kubectl", "run", "t", "--image=nginx", "--kubeconfig", kubeConfigPath)
	var cmdOut bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdOut
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(cmdOut.Bytes())
	}
}
func buildImage() {
	cmd := exec.Command("docker", "build", "-t", "watch-informer:ci", "..", "-f", "../Dockerfile")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(out.Bytes())
	}
}

func importImage() {
	cmd := exec.Command("kind", "load", "docker-image", "watch-informer:ci")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(out.Bytes())
		Expect(err).NotTo(HaveOccurred(), "Failed to import the Docker image into kind")
	}
}

func setupKindCluster() string {
	cmd := exec.Command("kind", "create", "cluster")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(out.Bytes())
		Expect(err).NotTo(HaveOccurred(), "Failed to create the kind cluster")
	}

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	return kubeConfigPath
}

func teardownKindCluster(kubeConfigPath string) {
	cmd := exec.Command("kind", "delete", "cluster", "--kubeconfig", kubeConfigPath)
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())

}

func deployApplication(kubeConfigPath string) {
	cmd := exec.Command("kubectl", "apply", "-k", "../kustomize/overlays/ci", "--kubeconfig", kubeConfigPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		GinkgoWriter.Write(out.Bytes())
		Expect(err).NotTo(HaveOccurred(), "Failed to deploy the application")
	}
}

func getPodLogs(kubeConfigPath string) string {
	cmd := exec.Command("kubectl", "logs", "deploy/watch-informer", "-n", "watch-informer", "--kubeconfig", kubeConfigPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
	return out.String()
}
