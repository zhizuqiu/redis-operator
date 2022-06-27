package redis_client

import (
	"bytes"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestCli(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		namespace     = "default"
		containerName = ""
		podName       = "rfr-redis-sample-0"
		command       = "redis-cli -a pass info"
	)

	// For now I am assuming stdin for the command to be nil
	output, stderr, err := ExecToPodThroughAPI(command, containerName, podName, namespace, nil)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr)
	}
	if err != nil {
		fmt.Printf("Error occured while `exec`ing to the State %q, namespace %q, command %q. Error: %+v\n", podName, namespace, command, err)
	} else {
		fmt.Println("Output:")
		fmt.Println(output)
	}
}

// ExecToPodThroughAPI uninterractively exec to the pod with the command specified.
// :param string command: list of the str which specify the command.
// :param string pod_name: State name
// :param string namespace: namespace of the State.
// :param io.Reader stdin: Standerd Input if necessary, otherwise `nil`
// :return: string: Output of the command. (STDOUT)
//          string: Errors. (STDERR)
//           error: If any error has occurred otherwise `nil`
func ExecToPodThroughAPI(command string, containerName, podName, namespace string, stdin io.Reader) (string, string, error) {
	config := ctrl.GetConfigOrDie()
	clientset, err := GetClientsetFromConfig(config)
	if err != nil {
		return "", "", err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return "", "", fmt.Errorf("error adding to scheme: %v", err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command: []string{
			"sh",
			"-c",
			command,
		},
		Container: containerName,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	// fmt.Println("Request URL:", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}

// GetClientsetFromConfig takes REST config and Create a clientset based on that and return that clientset
func GetClientsetFromConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = fmt.Errorf("failed creating clientset. Error: %+v", err)
		return nil, err
	}

	return clientset, nil
}

func TestEscapeRedisPassword(t *testing.T) {
	var decryptTests = []struct {
		in       string
		expected string
	}{
		{"HyxfHdIpiCui4jA", "HyxfHdIpiCui4jA"},
		{"HyxfHdIpiCui4j$A", "HyxfHdIpiCui4j\\$A"},
		{"$A", "\\$A"},
		{"$", "\\$"},
		{"", ""},
	}

	for _, tt := range decryptTests {
		actual := EscapeRedisPassword(tt.in)
		if actual != tt.expected {
			t.Errorf("EscapeRedisPassword(%s) = %s; expected %s", tt.in, actual, tt.expected)
		}
	}
}
