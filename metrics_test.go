package lib_test

import (
	"fmt"
	lib "git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"runtime"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// Тестируем ValidateNameVersion
func TestValidateNameVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		project, types, domain string
		wantName, wantVer      string
	}{
		{"", "", "", "unknown", ""},
		{"proj", "t", "", "proj", "t"},
		{"a-b-c-d", "ign", "dom/ver", "dom", "ver"},
		{"", "ty", "dom", "dom", "ty"},
		{"", "ty", "dom/ver", "dom", "ver"},
		{"svc", "", "svc/ver", "svc", "ver"},
	}

	for _, tc := range cases {
		name, ver := lib.ValidateNameVersion(tc.project, tc.types, tc.domain)
		assert.Equal(t, tc.wantName, name, fmt.Sprintf("project=%q types=%q domain=%q", tc.project, tc.types, tc.domain))
		assert.Equal(t, tc.wantVer, ver, fmt.Sprintf("project=%q types=%q domain=%q", tc.project, tc.types, tc.domain))
	}
}

// Тестируем NewBuildInfo + SetBuildInfo через собственное registry
func TestNewAndSetBuildInfo(t *testing.T) {
	t.Parallel()
	// 1) создаём новую GaugeVec
	gv := lib.NewBuildInfo("svcX")

	// 2) регистрируем во временном реестре
	reg := prometheus.NewRegistry()
	err := reg.Register(gv)
	assert.NoError(t, err)

	// 3) выставляем метрику
	lib.SetBuildInfo("v1", "revA", "dcZ")

	// 4) собираем все метрики
	mfs, err := reg.Gather()
	assert.NoError(t, err)

	// Ищем нужный MetricFamily
	var foundMF *dto.MetricFamily
	for _, mf := range mfs {
		if mf.GetName() == "svcX_build_info" {
			foundMF = mf
			break
		}
	}
	assert.NotNil(t, foundMF, "должен быть найден mf svcX_build_info")
	assert.Contains(t, foundMF.GetHelp(), "svcX was built")

	// Должен быть ровно один экземпляр метрики
	metrics := foundMF.GetMetric()
	assert.Len(t, metrics, 1)

	m := metrics[0]
	// Проверяем лейблы
	lbls := map[string]string{}
	for _, lp := range m.GetLabel() {
		lbls[lp.GetName()] = lp.GetValue()
	}
	assert.Equal(t, "v1", lbls["version"])
	assert.Equal(t, "revA", lbls["revision"])
	assert.Equal(t, "dcZ", lbls["dc"])
	assert.Equal(t, runtime.Version(), lbls["goversion"])
	// Значение = 1
	assert.Equal(t, float64(1), m.GetGauge().GetValue())
}
