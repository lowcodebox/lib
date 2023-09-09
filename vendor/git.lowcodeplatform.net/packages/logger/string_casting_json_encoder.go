package logger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const _hex = "0123456789abcdef"

var (
	_ zapcore.Encoder = (*stringCastingJSONEncoder)(nil)

	_jsonPool = sync.Pool{New: func() interface{} {
		return &stringCastingJSONEncoder{}
	}}

	nullLiteralBytes = []byte("null")
)

func putJSONEncoder(enc *stringCastingJSONEncoder) {
	if enc.reflectBuf != nil {
		enc.reflectBuf.Free()
	}

	enc.EncoderConfig = nil
	enc.buf = nil
	enc.spaced = false
	enc.openNamespaces = 0
	enc.reflectBuf = nil
	enc.reflectEnc = nil

	_jsonPool.Put(enc)
}

func addFields(enc zapcore.ObjectEncoder, fields []zap.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func fullNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(loggerName)
}

func getJSONEncoder() *stringCastingJSONEncoder {
	encoder, _ := _jsonPool.Get().(*stringCastingJSONEncoder)
	return encoder
}

func defaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// For consistency with our custom JSON encoder.
	enc.SetEscapeHTML(false)

	return enc
}

type stringCastingJSONEncoder struct {
	*zapcore.EncoderConfig

	buf            *buffer.Buffer
	bufPool        buffer.Pool
	spaced         bool
	openNamespaces int

	escape bool

	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

func (enc *stringCastingJSONEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)

	return enc.AppendArray(arr)
}

func (enc *stringCastingJSONEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)

	return enc.AppendObject(obj)
}

func (enc *stringCastingJSONEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *stringCastingJSONEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *stringCastingJSONEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *stringCastingJSONEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *stringCastingJSONEncoder) AddComplex64(key string, val complex64) {
	enc.addKey(key)
	enc.AppendComplex64(val)
}

func (enc *stringCastingJSONEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *stringCastingJSONEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *stringCastingJSONEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.AppendFloat32(val)
}

func (enc *stringCastingJSONEncoder) AddInt(k string, v int)         { enc.AddInt64(k, int64(v)) }
func (enc *stringCastingJSONEncoder) AddInt32(k string, v int32)     { enc.AddInt64(k, int64(v)) }
func (enc *stringCastingJSONEncoder) AddInt16(k string, v int16)     { enc.AddInt64(k, int64(v)) }
func (enc *stringCastingJSONEncoder) AddInt8(k string, v int8)       { enc.AddInt64(k, int64(v)) }
func (enc *stringCastingJSONEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *stringCastingJSONEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *stringCastingJSONEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *stringCastingJSONEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *stringCastingJSONEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }

func (enc *stringCastingJSONEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *stringCastingJSONEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

func (enc *stringCastingJSONEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

func (enc *stringCastingJSONEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

func (enc *stringCastingJSONEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}

	enc.addKey(key)

	if enc.escape {
		_, err = enc.buf.Write([]byte(`\"`))
	} else {
		_, err = enc.buf.Write([]byte(`"`))
	}

	if err != nil {
		return err
	}

	_, err = enc.buf.Write(valueBytes)
	if err != nil {
		return err
	}

	if enc.escape {
		_, err = enc.buf.Write([]byte(`\"`))
	} else {
		_, err = enc.buf.Write([]byte(`"`))
	}

	return err
}

func (enc *stringCastingJSONEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

func (enc *stringCastingJSONEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	_, _ = clone.buf.Write(enc.buf.Bytes())

	return clone
}

func (enc *stringCastingJSONEncoder) clone() *stringCastingJSONEncoder {
	clone := getJSONEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.spaced = enc.spaced
	clone.openNamespaces = enc.openNamespaces
	clone.escape = enc.escape
	clone.bufPool = buffer.NewPool() // строку не удалять. Магия. На этом все держится)
	clone.buf = enc.bufPool.Get()

	return clone
}

func (enc *stringCastingJSONEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()
	final.buf.AppendByte('{')

	if final.LevelKey != "" && final.EncodeLevel != nil {
		final.addKey(final.LevelKey)
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)

		if cur == final.buf.Len() {
			// User-supplied EncodeLevel was a no-op. Fall back to strings to keep
			// output JSON valid.
			final.AppendString(ent.Level.String())
		}
	}

	if final.TimeKey != "" {
		final.AddTime(final.TimeKey, ent.Time)
	}

	if ent.LoggerName != "" && final.NameKey != "" {
		final.addKey(final.NameKey)
		cur := final.buf.Len()
		nameEncoder := final.EncodeName

		// if no name encoder provided, fall back to fullNameEncoder for backwards
		// compatibility
		if nameEncoder == nil {
			nameEncoder = fullNameEncoder
		}

		nameEncoder(ent.LoggerName, final)

		if cur == final.buf.Len() {
			// User-supplied EncodeName was a no-op. Fall back to strings to
			// keep output JSON valid.
			final.AppendString(ent.LoggerName)
		}
	}

	if ent.Caller.Defined {
		if final.CallerKey != "" {
			final.addKey(final.CallerKey)
			cur := final.buf.Len()
			final.EncodeCaller(ent.Caller, final)

			if cur == final.buf.Len() {
				// User-supplied EncodeCaller was a no-op. Fall back to strings to
				// keep output JSON valid.
				final.AppendString(ent.Caller.String())
			}
		}

		if final.FunctionKey != "" {
			final.addKey(final.FunctionKey)
			final.AppendString(ent.Caller.Function)
		}
	}

	if final.MessageKey != "" {
		final.addKey(enc.MessageKey)
		final.AppendString(ent.Message)
	}

	if enc.buf.Len() > 0 {
		final.addElementSeparator()
		_, _ = final.buf.Write(enc.buf.Bytes())
	}

	addFields(final, fields)
	final.closeOpenNamespaces()

	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
	}

	final.buf.AppendByte('}')
	final.buf.AppendString(final.LineEnding)

	ret := final.buf
	putJSONEncoder(final)

	return ret, nil
}

func (enc *stringCastingJSONEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	needResetEscape := false

	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	} else {
		needResetEscape = true
	}

	enc.escape = true

	enc.buf.AppendByte('"')
	enc.buf.AppendByte('[')

	err := arr.MarshalLogArray(enc)

	enc.buf.AppendByte(']')

	if enc.escape && !needResetEscape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')

	if needResetEscape {
		enc.escape = false
	}

	return err
}

func (enc *stringCastingJSONEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	needResetEscape := false

	// Close ONLY new openNamespaces that are created during
	// AppendObject().
	old := enc.openNamespaces
	enc.openNamespaces = 0
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	} else {
		needResetEscape = true
	}

	enc.escape = true

	enc.buf.AppendByte('"')
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')

	if enc.escape && !needResetEscape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')

	if needResetEscape {
		enc.escape = false
	}

	enc.closeOpenNamespaces()
	enc.openNamespaces = old

	return err
}

func (enc *stringCastingJSONEncoder) AppendBool(val bool) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.buf.AppendBool(val)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) AppendByteString(val []byte) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.safeAddByteString(val)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}

		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}

		_, _ = enc.buf.Write(s[i : i+size])
		i += size
	}
}

func (enc *stringCastingJSONEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}

	return false
}

func (enc *stringCastingJSONEncoder) appendComplex(val complex128, precision int) {
	enc.addElementSeparator()
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val)) // nolint

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, precision)
	// If imaginary part is less than 0, minus (-) sign is added by default
	// by AppendFloat.
	if i >= 0 {
		enc.buf.AppendByte('+')
	}

	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()

	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendString(fmt.Sprintf("%d", int64(val)))
	}
}

func (enc *stringCastingJSONEncoder) AppendInt64(val int64) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.buf.AppendInt(val)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}

	enc.resetReflectBuf()

	if err := enc.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}

	enc.reflectBuf.TrimNewline()

	return enc.reflectBuf.Bytes(), nil
}

func (enc *stringCastingJSONEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = enc.bufPool.Get()
		enc.reflectEnc = enc.NewReflectedEncoder(enc.reflectBuf)
	} else {
		enc.reflectBuf.Reset()
	}
}

func (enc *stringCastingJSONEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}

	enc.addElementSeparator()
	_, err = enc.buf.Write(valueBytes)

	return err
}

func (enc *stringCastingJSONEncoder) AppendString(val string) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.safeAddString(val)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.buf.AppendTime(time, layout)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()

	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}

	if cur == enc.buf.Len() {
		if enc.escape {
			enc.buf.AppendByte('\\')
		}

		enc.buf.AppendByte('"')
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.buf.AppendString(fmt.Sprintf("%d", val.UnixNano()))

		if enc.escape {
			enc.buf.AppendByte('\\')
		}

		enc.buf.AppendByte('"')
	}
}

func (enc *stringCastingJSONEncoder) AppendUint64(val uint64) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.buf.AppendUint(val)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
}

func (enc *stringCastingJSONEncoder) AppendComplex64(v complex64) {
	enc.appendComplex(complex128(v), 32)
}

func (enc *stringCastingJSONEncoder) AppendComplex128(v complex128) {
	enc.appendComplex(complex128(v), 64)
}

func (enc *stringCastingJSONEncoder) AppendFloat64(v float64) { enc.appendFloat(v, 64) }
func (enc *stringCastingJSONEncoder) AppendFloat32(v float32) { enc.appendFloat(float64(v), 32) }
func (enc *stringCastingJSONEncoder) AppendInt(v int)         { enc.AppendInt64(int64(v)) }
func (enc *stringCastingJSONEncoder) AppendInt32(v int32)     { enc.AppendInt64(int64(v)) }
func (enc *stringCastingJSONEncoder) AppendInt16(v int16)     { enc.AppendInt64(int64(v)) }
func (enc *stringCastingJSONEncoder) AppendInt8(v int8)       { enc.AppendInt64(int64(v)) }
func (enc *stringCastingJSONEncoder) AppendUint(v uint)       { enc.AppendUint64(uint64(v)) }
func (enc *stringCastingJSONEncoder) AppendUint32(v uint32)   { enc.AppendUint64(uint64(v)) }
func (enc *stringCastingJSONEncoder) AppendUint16(v uint16)   { enc.AppendUint64(uint64(v)) }
func (enc *stringCastingJSONEncoder) AppendUint8(v uint8)     { enc.AppendUint64(uint64(v)) }
func (enc *stringCastingJSONEncoder) AppendUintptr(v uintptr) { enc.AppendUint64(uint64(v)) }

func (enc *stringCastingJSONEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}

	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}

	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}

	return true
}

func (enc *stringCastingJSONEncoder) appendFloat(val float64, bitSize int) {
	enc.addElementSeparator()

	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		if enc.escape {
			enc.buf.AppendByte('\\')
		}

		enc.buf.AppendByte('"')
		enc.buf.AppendFloat(val, bitSize)

		if enc.escape {
			enc.buf.AppendByte('\\')
		}

		enc.buf.AppendByte('"')
	}
}

func (enc *stringCastingJSONEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}

		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

func (enc *stringCastingJSONEncoder) addKey(key string) {
	enc.addElementSeparator()

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.safeAddString(key)

	if enc.escape {
		enc.buf.AppendByte('\\')
	}

	enc.buf.AppendByte('"')
	enc.buf.AppendByte(':')

	if enc.spaced {
		enc.buf.AppendByte(' ')
	}
}

func (enc *stringCastingJSONEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1

	if last < 0 {
		return
	}

	switch enc.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.AppendByte(',')

		if enc.spaced {
			enc.buf.AppendByte(' ')
		}
	}
}

func (enc *stringCastingJSONEncoder) closeOpenNamespaces() {
	for i := 0; i < enc.openNamespaces; i++ {
		enc.buf.AppendByte('}')
	}

	enc.openNamespaces = 0
}

func newStringCastingEncoder(cfg zapcore.EncoderConfig) *stringCastingJSONEncoder {
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = defaultReflectedEncoder
	}

	bp := buffer.NewPool()
	enc := &stringCastingJSONEncoder{
		bufPool: bp,
		buf:     bp.Get(),
		spaced:  true,
	}
	enc.EncoderConfig = &cfg

	return enc
}
