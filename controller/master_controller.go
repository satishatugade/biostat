package controller

import (
	"biostat/service"
)

type MasterController struct {
	allergyService service.AllergyService
}

func NewMasterController(allergyService service.AllergyService) *MasterController {
	return &MasterController{allergyService}
}
