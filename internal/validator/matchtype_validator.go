package validator

import (
	"fmt"
	"github.com/lccxxo/bailuoli/internal/model"
)

type MatchTypeValidator struct {
	BaseValidator
}

func (v *MatchTypeValidator) Validate(route *model.Route) error {
	switch route.MatchType {
	case "exact", "prefix", "regex":
		if v.next != nil {
			return v.next.Validate(route)
		}
		return nil
	default:
		return fmt.Errorf("invalid match type: %s", route.MatchType)
	}
}
