package validator

import (
	"fmt"
	"github.com/lccxxo/bailuoli/internal/model"
)

type PathValidator struct {
	BaseValidator
}

func (v *PathValidator) Validate(route *model.Route) error {
	if route.Path == "" {
		return fmt.Errorf("route %s: path cannot be empty", route.Name)
	}
	if v.next != nil {
		return v.next.Validate(route)
	}
	return nil
}
