package validator

import (
	"net/http"
	"strings"
	"tickets-hunter/common/xerr"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
)

type CustomValidator struct {
	validate *validator.Validate
}

func New() *CustomValidator {
	v := validator.New()

	return &CustomValidator{
		validate: v,
	}
}

// Validate 实现 httpx.Validator 接口
func (v *CustomValidator) Validate(r *http.Request, data interface{}) error {
	// 调用 validator/v10 进行验证
	err := v.validate.Struct(data)
	if err != nil {
		// 这里的err是validator.ValidationErrors 类型
		// 解析错误信息，构造一个更友好的错误消息
		if errs, ok := err.(validator.ValidationErrors); ok {
			var msgs []string
			for _, err := range errs {
				// err.Field() 是结构体字段名，err.Tag() 是校验规则
				// 简单的格式化： "Nickname is required"
				// 你也可以根据 err.Tag() 做更详细的映射
				msgs = append(msgs, err.Field()+" should confirm "+err.Tag())
			}
			// 用逗号拼接所有错误消息，并返回一个自定义的 ValidationError
			return xerr.NewValidationError(uint32(codes.InvalidArgument), strings.Join(msgs, ", "))
		} else {
			// 如果不是 ValidationErrors 类型，直接返回原始错误，将来全局错误处理器会将其作为系统错误处理：打印日志，并且返回一个通用的错误消息给前端
			return err
		}
	}
	return nil
}

// Engine 暴露内部验证器，以便在其他地方使用
func (v *CustomValidator) Engine() *validator.Validate {
	return v.validate
}
