package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/util/sm4"
	corev1 "k8s.io/api/core/v1"
	apl "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSpecRedisPassword retreives password from kubernetes secret or, if
// unspecified, returns a blank string
func GetSpecRedisPassword(s Services, rf *roav1.Redis) (string, error) {

	if rf.Spec.Auth.SecretPath != "" {
		secret, err := s.GetSecret(rf.Namespace, rf.Spec.Auth.SecretPath)
		if err != nil {
			return "", err
		}

		if password, ok := secret.Data["password"]; ok {
			return string(password), nil
		}

		return "", fmt.Errorf("secret \"%s\" does not have a password field", rf.Spec.Auth.SecretPath)
	} else {
		if rf.Spec.Auth.Password.Value != "" {
			switch rf.Spec.Auth.Password.EncodeType {
			case roav1.BASE64:
				passwordByte, err := base64.StdEncoding.DecodeString(rf.Spec.Auth.Password.Value)
				if err != nil {
					return "", err
				}
				return string(passwordByte), nil
			case roav1.SM4:
				password, err := sm4.DecryptSm4([]byte(rf.Spec.Auth.Password.Value), sm4.Sm4Key)
				if err != nil {
					return "", err
				}
				return password, nil
			default:
				passwordByte, err := base64.StdEncoding.DecodeString(rf.Spec.Auth.Password.Value)
				if err != nil {
					return "", err
				}
				return string(passwordByte), nil
			}
		}
	}

	return "", nil
}

func ListPods(kubeClient client.Client, namespace string, selector map[string]string) (*corev1.PodList, error) {
	var podList = &corev1.PodList{}
	if err := kubeClient.List(context.Background(),
		podList,
		// client.InNamespace(req.Namespace),
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: apl.SelectorFromSet(selector),
		},
		// client.MatchingFields{ownerKey: req.Name},
	); err != nil {
		return nil, err
	}
	return podList, nil
}
