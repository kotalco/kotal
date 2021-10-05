package shared

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestSecurityContext(t *testing.T) {
	var userId int64 = 1000
	var groupId int64 = 3000
	var fsGroupId int64 = 2000
	var nonRoot = true
	policy := corev1.FSGroupChangeOnRootMismatch

	context := SecurityContext()

	if *context.RunAsUser != userId {
		t.Errorf("expected user id to be %d, got %d", userId, *context.RunAsUser)
	}

	if *context.RunAsGroup != groupId {
		t.Errorf("expected group id to be %d, got %d", groupId, *context.RunAsGroup)
	}

	if *context.FSGroup != fsGroupId {
		t.Errorf("expected fs group id to be %d, got %d", fsGroupId, *context.FSGroup)
	}

	if *context.RunAsNonRoot != nonRoot {
		t.Errorf("expected non root to be %t, got %t", nonRoot, *context.RunAsNonRoot)
	}

	if *context.FSGroupChangePolicy != policy {
		t.Errorf("expected fs group change policy to be %s, got %s", policy, *context.FSGroupChangePolicy)
	}

}
