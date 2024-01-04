package app_lib

import (
	"fmt"
	"testing"

	"git.lowcodeplatform.net/fabric/lib"
)

var config struct {

	// VFS
	VfsBucket      string `envconfig:"VFS_BUCKET" default:"buildbox"`
	VfsKind        string `envconfig:"VFS_KIND" default:"s3"`
	VfsEndpoint    string `envconfig:"VFS_ENDPOINT" default:"http://127.0.0.1:9000"`
	VfsAccessKeyId string `envconfig:"VFS_ACCESS_KEY_ID" default:"minioadmin"`
	VfsSecretKey   string `envconfig:"VFS_SECRET_KEY" default:"minioadmin"`
	VfsRegion      string `envconfig:"VFS_REGION" default:""`
	VfsComma       string `envconfig:"VFS_COMMA" default:""`
	VfsCertCA      string `envconfig:"VFS_CERT_CA" default:""`
}

func Test_csvtosliсemap(t *testing.T) {
	in := "field1;field2\n2;3"

	NewFuncMap(nil, nil)
	res, err := Funcs.csvtosliсemap([]byte(in))
	if err != nil {
		t.Errorf("Should not produce an error")
	}

	if res[0]["field1"] != "2" {
		t.Errorf("Result was incorrect, got: %s, want: %s.", res[0]["field1"], "2")
	}
}

func Test_unzip(t *testing.T) {
	cfg := config
	cfg.VfsBucket = "buildbox"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	in := "logo_ep_l_g.png"

	NewFuncMap(vfs, nil)
	status := Funcs.unzip(in, "")

	fmt.Println(status)
}

func Test_imgResize(t *testing.T) {
	cfg := config
	cfg.VfsBucket = "buildbox"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	in := "landing/ludam.png"

	NewFuncMap(vfs, nil)

	res := Funcs.imgResize(in, 100, 100)

	fmt.Println("result:", res)
}

func Test_imgCrop(t *testing.T) {
	cfg := config
	cfg.VfsBucket = "buildbox"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	in := "landing/katya.jpg"

	NewFuncMap(vfs, nil)

	res := Funcs.imgCrop(in, 500, 500, true, false, 0, 0)

	fmt.Println("result:", res)
}

func Test_imgCropAndResize(t *testing.T) {
	cfg := config
	cfg.VfsBucket = "buildbox"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	in := "landing/katya.jpg"

	NewFuncMap(vfs, nil)

	res := Funcs.imgCrop(in, 500, 500, true, false, 0, 0)
	res = Funcs.imgResize(res, 100, 100)

	fmt.Println("result:", res)
}
