package validation

// Defaults 参数默认值
type Defaults struct {
	values map[string]interface{}
}

// NewDefaults 创建默认值配置
func NewDefaults(values map[string]interface{}) *Defaults {
	copyValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		copyValues[k] = v
	}
	return &Defaults{values: copyValues}
}

// Add 添加默认值，如果键已存在则保持原值
func (d *Defaults) Add(key string, value interface{}) {
	if d == nil {
		return
	}
	if _, exists := d.values[key]; !exists {
		d.values[key] = value
	}
}

// Set 设置默认值，覆盖已存在的键
func (d *Defaults) Set(key string, value interface{}) {
	if d == nil {
		return
	}
	d.values[key] = value
}

// Get 获取默认值
func (d *Defaults) Get(key string) (interface{}, bool) {
	if d == nil {
		return nil, false
	}
	val, ok := d.values[key]
	return val, ok
}

// Merge 合并默认值配置，其他配置优先级更高
func (d *Defaults) Merge(other *Defaults) {
	if d == nil || other == nil {
		return
	}
	for k, v := range other.values {
		if _, exists := d.values[k]; !exists {
			d.values[k] = v
		}
	}
}

// isZeroValue 判断参数是否为“零值”或未设置值。
func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0
	case float32:
		return v == 0
	default:
		return false
	}
}

// Apply 应用默认值
func (d *Defaults) Apply(params map[string]interface{}) {
	if d == nil {
		return
	}
	for k, v := range d.values {
		if existing, exists := params[k]; !exists || isZeroValue(existing) {
			params[k] = v
		}
	}
}

// PageDefaults 分页参数默认值
var PageDefaults = NewDefaults(map[string]interface{}{
	"page":        1,
	"size":        50,
	"order_field": "created_at",
	"order_type":  "desc",
})

// UploadOptionsDefaults 上传选项默认值
var UploadOptionsDefaults = NewDefaults(map[string]interface{}{
	"policy": "skip",
})
