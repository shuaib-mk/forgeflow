package authorization

import (
	"github.com/forgeflow/forgeflow/internal/domain"
	"github.com/forgeflow/forgeflow/pkg/models"
)

type Permission string
const (Read Permission="read";Write Permission="write";Administer Permission="administer")

func Require(role models.Role,permission Permission)error{
	allowed:=map[models.Role]map[Permission]bool{
		models.RoleOwner:{Read:true,Write:true,Administer:true},
		models.RoleAdmin:{Read:true,Write:true,Administer:true},
		models.RoleDeveloper:{Read:true,Write:true},
		models.RoleViewer:{Read:true},
	}
	if !allowed[role][permission]{return domain.ErrForbidden};return nil
}

