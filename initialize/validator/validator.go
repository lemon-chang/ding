package validator

import (
	"ding/global"
	"github.com/go-playground/validator/v10"
)

func Init() {
	global.GLOAB_VALIDATOR = validator.New()
	global.GLOAB_VALIDATOR.RegisterValidation("secret_required_if_type_2", secretRequiredIfType2)

}
func secretRequiredIfType2(fl validator.FieldLevel) bool {
	typeField := fl.Parent().FieldByName("Type")
	secretField := fl.Field().String()

	if typeField.String() == "2" {
		return len(secretField) == 67
	} else if typeField.String() == "1" {
		return len(secretField) == 0
	}

	return true
}
