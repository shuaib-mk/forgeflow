package authorization

import (
	"github.com/forgeflow/forgeflow/pkg/models"
	"testing"
)

func TestAuthorizationMatrix(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		role       models.Role
		permission Permission
		allowed    bool
	}{
		{models.RoleOwner, Administer, true}, {models.RoleAdmin, Administer, true}, {models.RoleDeveloper, Write, true}, {models.RoleDeveloper, Administer, false}, {models.RoleViewer, Read, true}, {models.RoleViewer, Write, false},
	} {
		if got := Require(tt.role, tt.permission) == nil; got != tt.allowed {
			t.Errorf("Require(%s,%s) allowed=%v", tt.role, tt.permission, got)
		}
	}
}
