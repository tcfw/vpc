package hyper

import (
	"encoding/xml"
	"fmt"
	"strings"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
)

func GetTemplate(vm *hyperAPI.VM) (*DomainDesc, error) {
	if strings.TrimSpace(vm.TemplateId)[0] == '<' {
		desc := &DomainDesc{}
		desc.Name = vm.Id

		if err := xml.Unmarshal([]byte(vm.TemplateId), desc); err != nil {
			return nil, fmt.Errorf("failed to marshal desc: %s", err)
		}

		return desc, nil
	}

	return nil, fmt.Errorf("unknown template reference")
}
