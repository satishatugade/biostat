package constant

import (
	"errors"
	"fmt"
)

type ErrorManager struct {
	GenericErrors map[string]string
}

func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		GenericErrors: map[string]string{
			"Success":          "%s deleted succesfully Id: %d",
			"NotFound":         "%s not found with given Id: %d",
			"DeleteError":      "Failed to delete %s with Id: %d",
			"UpdateError":      "Failed to update %s with Id: %d",
			"AuditError":       "Failed to save %s audit: %v",
			"NoRowsAffected":   "No rows affected for %s with Id: %d",
			"TransactionError": "An unexpected error occured %s ",
			"CreateError":      "Error while creating %s",
		},
	}
}

func (em *ErrorManager) ErrorMessage(errorType string, args ...interface{}) error {
	if msg, ok := em.GenericErrors[errorType]; ok {
		if len(args) > 0 {
			return fmt.Errorf(msg, args...)
		}
		return errors.New(msg)
	}
	return fmt.Errorf("error message not defined for errorType: %s", errorType)
}
