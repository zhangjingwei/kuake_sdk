package validation

// ValidatorFunc 参数校验函数类型
type ValidatorFunc func(value interface{}) error

// Validator 参数校验器
type Validator struct {
	name     string
	validate ValidatorFunc
}

// NewValidator 创建新的校验器
func NewValidator(name string, fn ValidatorFunc) *Validator {
	return &Validator{name: name, validate: fn}
}

// Validate 执行校验
func (v *Validator) Validate(value interface{}) error {
	return v.validate(value)
}

// Chain 链式校验
func Chain(validators ...*Validator) ValidatorFunc {
	return func(value interface{}) error {
		for _, v := range validators {
			if err := v.Validate(value); err != nil {
				return err
			}
		}
		return nil
	}
}
