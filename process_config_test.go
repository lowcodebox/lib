package lib_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	lib "git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {

	base := lib.ProcessConfig{
		Path:    "p",
		Project: "proj",
		Service: "svc",
		Config:  "cfg",
		Command: "cmd",
		Mode:    "m",
		DC:      "dc",
		Port:    "8080",
	}
	cases := []struct {
		name    string
		modify  func(pc *lib.ProcessConfig)
		wantErr string
	}{
		{
			"empty Path",
			func(pc *lib.ProcessConfig) { pc.Path = "" },
			"path cannot be empty",
		},
		{
			"empty Project",
			func(pc *lib.ProcessConfig) { pc.Project = "" },
			"project cannot be empty",
		},
		{
			"empty Service",
			func(pc *lib.ProcessConfig) { pc.Service = "" },
			"service cannot be empty",
		},
		{
			"empty Config",
			func(pc *lib.ProcessConfig) { pc.Config = "" },
			"config cannot be empty",
		},
		{
			"empty Command",
			func(pc *lib.ProcessConfig) { pc.Command = "" },
			"command cannot be empty",
		},
		{
			"empty Mode",
			func(pc *lib.ProcessConfig) { pc.Mode = "" },
			"mode cannot be empty",
		},
		{
			"empty DC",
			func(pc *lib.ProcessConfig) { pc.DC = "" },
			"dc cannot be empty",
		},
		{
			"empty Port",
			func(pc *lib.ProcessConfig) { pc.Port = "" },
			"port cannot be empty",
		},
		{
			"all ok",
			func(pc *lib.ProcessConfig) {},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pc := base
			tc.modify(&pc)
			err := pc.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestBuildArgs(t *testing.T) {

	pc := lib.ProcessConfig{
		Command: "run",
		Config:  "conf.toml",
		Mode:    "prod",
		DC:      "eu",
		Port:    "9000",
	}
	args := pc.BuildArgs()
	assert.Equal(t, []string{"run", "-c", "conf.toml", "-m", "prod", "-d", "eu", "-p", "9000"}, args)
}

func TestString_GetLogFileName_GetAuditInfo(t *testing.T) {

	pc := lib.ProcessConfig{
		Path:    "/bin/app",
		Project: "myproj",
		Service: "svc",
		Config:  "/etc/cfg",
		Command: "run",
		Mode:    "debug",
		DC:      "dc1",
		Port:    "1234",
	}
	s := pc.String()
	// должен содержать ключевые поля
	assert.Contains(t, s, "Path: /bin/app")
	assert.Contains(t, s, "Project: myproj")
	assert.Contains(t, s, "Service: svc")

	assert.Equal(t, "myproj-svc.log", pc.GetLogFileName())

	audit := pc.GetAuditInfo()
	expectMap := map[string]string{
		"path":    "/bin/app",
		"project": "myproj",
		"service": "svc",
		"config":  "/etc/cfg",
		"command": "run",
		"mode":    "debug",
		"dc":      "dc1",
		"port":    "1234",
	}
	assert.Equal(t, expectMap, audit)
}

func TestClone(t *testing.T) {

	orig := &lib.ProcessConfig{
		Path:    "/a",
		Project: "p",
		Service: "s",
		Config:  "c",
		Command: "x",
		Mode:    "m",
		DC:      "d",
		Port:    "80",
	}
	clone := orig.Clone()
	// разные указатели
	assert.NotSame(t, orig, clone)
	// но значения равны
	assert.Equal(t, orig, clone)

	// изменение клона не затрагивает оригинал
	clone.Service = "other"
	assert.NotEqual(t, orig.Service, clone.Service)
}

func TestValidateAndSanitize_Success(t *testing.T) {

	tmp := t.TempDir()
	// создаём «исполняемый» файл и конфиг
	execPath := filepath.Join(tmp, "exe-file")
	configPath := filepath.Join(tmp, "cfg-file")
	assert.NoError(t, os.WriteFile(execPath, []byte{}, 0755))
	assert.NoError(t, os.WriteFile(configPath, []byte{}, 0644))

	pc := lib.ProcessConfig{
		Path:    "  " + execPath + "  ",
		Config:  "\t" + configPath + "\n",
		Project: " proj ",
		Service: " svc ",
		Command: " cmd ",
		Mode:    " mode ",
		DC:      " dc ",
		Port:    " 42 ",
	}

	err := pc.ValidateAndSanitize()
	assert.NoError(t, err)

	// должны быть обрезаны пробелы
	assert.Equal(t, execPath, pc.Path)
	assert.Equal(t, configPath, pc.Config)
	assert.Equal(t, "proj", pc.Project)
	assert.Equal(t, "svc", pc.Service)
	assert.Equal(t, "cmd", pc.Command)
	assert.Equal(t, "mode", pc.Mode)
	assert.Equal(t, "dc", pc.DC)
	assert.Equal(t, "42", pc.Port)
}

func TestValidateAndSanitize_InvalidExecPath(t *testing.T) {

	tmp := t.TempDir()
	// только конфиг существует
	configPath := filepath.Join(tmp, "cfg")
	assert.NoError(t, os.WriteFile(configPath, []byte{}, 0644))

	pc := lib.ProcessConfig{
		Path:   filepath.Join(tmp, "no-such-exe"),
		Config: configPath,
		// прочие обязательные поля
		Project: "p", Service: "s", Command: "c", Mode: "m", DC: "dc", Port: "1",
	}

	err := pc.ValidateAndSanitize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executable path does not exist")
}

func TestValidateAndSanitize_InvalidConfig(t *testing.T) {

	tmp := t.TempDir()
	// только exec существует
	execPath := filepath.Join(tmp, "exe")
	assert.NoError(t, os.WriteFile(execPath, []byte{}, 0755))

	pc := lib.ProcessConfig{
		Path:    execPath,
		Config:  filepath.Join(tmp, "no-cfg"),
		Project: "p", Service: "s", Command: "c", Mode: "m", DC: "dc", Port: "1",
	}

	err := pc.ValidateAndSanitize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file does not exist")
}

func TestGetCommandLine(t *testing.T) {

	pc := lib.ProcessConfig{
		Path:    "/bin/runme",
		Command: "run",
		Config:  "cfg",
		Mode:    "m",
		DC:      "dc",
		Port:    "7",
	}
	cmdline := pc.GetCommandLine()
	// должно начатьcя с Path
	assert.True(t, strings.HasPrefix(cmdline, "/bin/runme "))
	// и содержать остальные аргументы
	assert.Contains(t, cmdline, "-c cfg")
	assert.Contains(t, cmdline, "-m m")
	assert.Contains(t, cmdline, "-d dc")
	assert.Contains(t, cmdline, "-p 7")
}
