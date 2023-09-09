package types

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type where struct {
	FuncName string
	Path     string
	Line     int
	Host     string
	Project  string
}

func (w where) method() string {
	return w.FuncName
}

func (w where) project() string {
	return filepath.Join(w.Host, w.Project)
}

func (w where) path() string {
	return fmt.Sprintf("%s:%d", w.Path, w.Line)
}

func (w where) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	data := map[string]string{
		"method":  w.method(),
		"project": w.project(),
		"path":    w.path(),
	}

	for name, value := range data {
		if value != "" {
			enc.AddString(name, value)
		}
	}

	return nil
}

var (
	srcMark    = filepath.Join("go", "src")
	modMark    = filepath.Join("go", "pkg", "mod")
	goPathMark = filepath.Join("gopath", "src")
)

const separator = string(os.PathSeparator)

func WhereWithDeep(deep int) zap.Field {
	field := zap.Skip()

	var (
		w  where
		pc uintptr
		ok bool
	)

	pc, w.Path, w.Line, ok = runtime.Caller(deep)
	if !ok {
		return field
	}

	w.Path = strings.ReplaceAll(w.Path, separator, "/")

	var ix int
	if ix = strings.Index(w.Path, srcMark); ix > -1 {
		w.Path = w.Path[ix+len(srcMark)+1:]
	} else if ix = strings.Index(w.Path, modMark); ix > -1 {
		w.Path = w.Path[ix+len(modMark)+1:]
	} else if ix = strings.Index(w.Path, goPathMark); ix > -1 {
		w.Path = w.Path[ix+len(goPathMark)+1:]
	}

	funcDetails := runtime.FuncForPC(pc)
	if funcDetails == nil {
		return field
	}

	funcFQN := funcDetails.Name()

	funcParts := strings.Split(funcFQN, separator)
	if len(funcParts) < 3 {
		return field
	}

	w.FuncName = filepath.Join(funcParts[len(funcParts)-3:]...)

	if !strings.HasPrefix(w.Path, "/") {
		urlData, err := url.Parse("//" + w.Path)
		if err != nil {
			return field
		}

		if len(urlData.Path) == 0 {
			return field
		}

		w.Path = urlData.Path[1:]

		pathParts := strings.Split(w.Path, separator)
		if len(pathParts) < 2 {
			return field
		}

		w.Host = urlData.Host
		w.Project = filepath.Join(pathParts[0:2]...)
	}

	return zapcore.Field{
		Type:      zapcore.InlineMarshalerType,
		Interface: w,
	}
}

func Where() zap.Field {
	return WhereWithDeep(2)
}
