package logger

import "go.uber.org/zap"

type fieldType int

const (
	fieldTypeString fieldType = iota
	fieldTypeUint64
	fieldTypeFloat64
)

// Field - alias типов логов для работы с fieldStorage.
// В дальнейшем надо бы использовать для самого пакета логера, чтобы абстрагироваться от внешнего пакета zap.
type Field struct {
	typeOf fieldType
	key    string
	value  interface{}
}

// External - каст типов под внешний логер. Если будут добавлены новые типы для обработки, надо расширять кастинг,
// по аналогии с FieldUint64.
func (f *Field) External() zap.Field {
	switch f.typeOf {
	case fieldTypeString:
		{
			if v, ok := f.value.(string); ok {
				return zap.String(f.key, v)
			}
		}
	case fieldTypeUint64:
		{
			if v, ok := f.value.(uint64); ok {
				return zap.Uint64(f.key, v)
			}
		}
	case fieldTypeFloat64:
		{
			if v, ok := f.value.(float64); ok {
				return zap.Float64(f.key, v)
			}
		}
	}

	return zap.String("warn", "Field type is not found")
}

func FieldString(key string, val string) Field {
	return Field{
		typeOf: fieldTypeString,
		key:    key,
		value:  val,
	}
}

func FieldUint64(key string, val uint64) Field {
	return Field{
		typeOf: fieldTypeUint64,
		key:    key,
		value:  val,
	}
}

func FieldFloat64(key string, val float64) Field {
	return Field{
		typeOf: fieldTypeFloat64,
		key:    key,
		value:  val,
	}
}
