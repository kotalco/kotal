package shared

import corev1 "k8s.io/api/core/v1"

// SecurityContext is the pod security policy used by all containers
func SecurityContext() *corev1.PodSecurityContext {
	var userId int64 = 1000
	var groupId int64 = 3000
	var fsGroupId int64 = 2000
	var nonRoot = true
	policy := corev1.FSGroupChangeOnRootMismatch

	return &corev1.PodSecurityContext{
		RunAsUser:           &userId,
		RunAsGroup:          &groupId,
		RunAsNonRoot:        &nonRoot,
		FSGroup:             &fsGroupId,
		FSGroupChangePolicy: &policy,
	}
}
