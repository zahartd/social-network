package app

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/zahartd/social-network/src/services/user-service/internal/utils"
)

// RegisterValidators hooks custom funcs into gin‑validator once.
func RegisterValidators(_ *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			s, _ := fl.Field().Interface().(string)
			return utils.ValidatePhone(s)
		})
		_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			s, _ := fl.Field().Interface().(string)
			return utils.ValidatePassword(s)
		})
		_ = v.RegisterValidation("login", func(fl validator.FieldLevel) bool {
			s, _ := fl.Field().Interface().(string)
			return utils.ValidateLogin(s)
		})
	}
}
