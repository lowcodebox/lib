package app_lib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
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

	// LOGBOX
	LogboxEndpoint       string        `envconfig:"LOGBOX_ENDPOINT" default:"http://127.0.0.1:8999"`
	LogboxAccessKeyId    string        `envconfig:"LOGBOX_ACCESS_KEY_ID" default:""`
	LogboxSecretKey      string        `envconfig:"LOGBOX_SECRET_KEY" default:""`
	LogboxRequestTimeout time.Duration `envconnfig:"LOGBOX_REQUEST_TIMEOUT" default:"300ms"`

	// LOGBOX-client CircuitBreaker
	CbMaxRequestsLogbox uint32        `envconfig:"CB_MAX_REQUESTS_LOGBOX" default:"3" description:"максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии"`
	CbTimeoutLogbox     time.Duration `envconfig:"CB_TIMEOUT_LOGBOX" default:"5s" description:"период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние"`
	CbIntervalLogbox    time.Duration `envconfig:"CB_INTERVAL_LOGBOX" default:"5s" description:"циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков"`
}

func Test_csvtosliсemap(t *testing.T) {
	in := "field1;field2\n2;3"

	NewFuncMap(nil, nil, "")
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
	cfg.VfsBucket = "lms"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""

	// подключаемся к файловому хранилищу
	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)
	in := "WMS.zip"

	NewFuncMap(vfs, nil, "")
	status := Funcs.unzip(in, "")

	fmt.Println(status)
}

func Test_parsescorm(t *testing.T) {
	cfg := config
	cfg.VfsBucket = "lms"
	cfg.VfsKind = "s3"
	cfg.VfsEndpoint = "http://127.0.0.1:9000"
	cfg.VfsAccessKeyId = "minioadmin"
	cfg.VfsSecretKey = "minioadmin"
	cfg.VfsRegion = ""
	cfg.VfsComma = ""
	cfg.VfsCertCA = ""
	in := "WMS.zip"

	vfs := lib.NewVfs(cfg.VfsKind, cfg.VfsEndpoint, cfg.VfsAccessKeyId, cfg.VfsSecretKey, cfg.VfsRegion, cfg.VfsBucket, cfg.VfsComma, cfg.VfsCertCA)

	NewFuncMap(vfs, nil, "")
	index := Funcs.parsescorm(in, "")
	fmt.Printf("index: %s", index)
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

	NewFuncMap(vfs, nil, "")

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

	NewFuncMap(vfs, nil, "")

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

	NewFuncMap(vfs, nil, "")

	res := Funcs.imgCrop(in, 500, 500, true, false, 0, 0)
	res = Funcs.imgResize(res, 100, 100)

	fmt.Println("result:", res)
}

func Test_sliceuint8delete(t *testing.T) {
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

	in := []uint8{1, 2, 3, 4, 5, 6}

	NewFuncMap(vfs, nil, "")

	res := Funcs.sliceuint8delete(in, 2)

	fmt.Println("result:", res)
}

func Test_sortbyfield(t *testing.T) {
	in := `{
  "data": [
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7__datecreate",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2023-12-27 12:31:37.276831817 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7__groups",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7__owner",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "status": "",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "Татьяна Медведева"
        },
        "_title": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7__title",
          "editor": "",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл PDF аака"
        },
        "access_admin": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_block",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "description": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_description",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "акак"
        },
        "label": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_label",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_order",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "1"
        },
        "preview": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_preview",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "title": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_title",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "аака"
        },
        "to_build": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-12-27T12-26-04Z-1a4ea7_type",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
          "src": "tpl_lms_material_type_pdf",
          "status": "",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл PDF"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2023-12-27 12:31:37.5065968 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл PDF аака",
      "type": "",
      "uid": "2023-12-27T12-26-04Z-1a4ea7"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4__datecreate",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2023-12-27 14:45:34.925146387 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4__groups",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4__owner",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "status": "",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "Татьяна Медведева"
        },
        "_title": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4__title",
          "editor": "",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл PDF к"
        },
        "access_admin": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_block",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "description": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_description",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "file_pdf": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_file_pdf",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "/lms/gui/materials/pdf/WBPro_Описание.pdf"
        },
        "label": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_label",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_order",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "1"
        },
        "preview": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_preview",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "title": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_title",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "к"
        },
        "to_build": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-12-27T14-45-05Z-5521d4_type",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл PDF"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2023-12-27 14:45:35.121212443 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл PDF к",
      "type": "",
      "uid": "2023-12-27T14-45-05Z-5521d4"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-02T12-00-20Z-519412__datecreate",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "2023-11-02 12:02:15.438085039 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-02T12-00-20Z-519412__groups",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-02T12-00-20Z-519412__owner",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "Татьяна Медведева"
        },
        "_title": {
          "Uuid": "2023-11-02T12-00-20Z-519412__title",
          "editor": "",
          "rev": "2024-01-25 09:13:23.421725193 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Rfrf"
        },
        "access_admin": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-02T12-00-20Z-519412_block",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2024-01-25 09:13:23.399865252 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "description": {
          "Uuid": "2023-11-02T12-00-20Z-519412_description",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "fddfbf df bf df bdfbdf"
        },
        "label": {
          "Uuid": "2023-11-02T12-00-20Z-519412_label",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": ""
        },
        "start_url_scorm": {
          "Uuid": "2023-11-02T12-00-20Z-519412_start_url_scorm",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "/scormcontent/index.html"
        },
        "title": {
          "Uuid": "2023-11-02T12-00-20Z-519412_title",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "Rfrf"
        },
        "to_build": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-02T12-00-20Z-519412_type",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-02T12-00-20Z-519412_zip-scorm",
          "editor": "",
          "rev": "2023-11-03 08:01:42.920444793 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "scorm/scormAndA.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 09:13:23.421725193 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Rfrf",
      "type": "",
      "uid": "2023-11-02T12-00-20Z-519412"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c__datecreate",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "2023-11-08 09:28:15.855252084 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c__groups",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c__owner",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "Татьяна Медведева"
        },
        "_title": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c__title",
          "editor": "",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Без аналитики?? Склад WILDBERRIES глазами работника"
        },
        "access_admin": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_block",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "2023-10-26T09-24-36Z-dee236",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "2023-09-05T17-32-54Z-0e93d3",
          "value": "Склад WILDBERRIES глазами работника"
        },
        "description": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_description",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "уауа"
        },
        "label": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_label",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": ""
        },
        "preview": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_preview",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": ""
        },
        "start_url_scorm": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_start_url_scorm",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "/scormcontent/index.html"
        },
        "title": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_title",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "Без аналитики??"
        },
        "to_build": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_type",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-08T09-25-11Z-b87e5c_zip-scorm",
          "editor": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
          "src": "",
          "status": "000684e7-5bfa-305c-9ec4-f5b69b3d3417",
          "tpls": "",
          "value": "/scorm/priemka_tovara.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2023-11-09 13:56:16.945057115 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Без аналитики?? Склад WILDBERRIES глазами работника",
      "type": "",
      "uid": "2023-11-08T09-25-11Z-b87e5c"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554__datecreate",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "2023-12-07 13:15:07.825139929 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554__groups",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554__owner",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "WBProf Оператор"
        },
        "_title": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554__title",
          "editor": "",
          "rev": "2024-01-25 16:27:12.292962393 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Электронный курс «Обезличивание товара»"
        },
        "access_admin": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_block",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "2023-11-13T10-50-45Z-f57031",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "Процесс приёмки товаров на складе"
        },
        "description": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_description",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 14:35:15.801684966 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "На данной странице электронного курса собраны все полезные материалы по процессу «Обезличивание товара»"
        },
        "label": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_label",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 14:35:20.37801709 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_order",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:27:12.222303365 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "6"
        },
        "preview": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_preview",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "title": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_title",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 14:35:22.128310284 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Электронный курс «Обезличивание товара»"
        },
        "to_build": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_type",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-12-07T13-14-36Z-ab8554_zip-scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:16:42.602222888 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "scorm/obezlichivanie.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 16:27:12.292962393 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Электронный курс «Обезличивание товара»",
      "type": "",
      "uid": "2023-12-07T13-14-36Z-ab8554"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d__datecreate",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "2023-11-28 13:14:26.836926055 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d__groups",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d__owner",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "WBProf Оператор"
        },
        "_title": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d__title",
          "editor": "",
          "rev": "2024-01-25 16:30:09.175300894 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Электронный курс «Сборка товара»"
        },
        "access_admin": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_block",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "2023-11-13T10-51-08Z-db462b",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "Процесс сборки товаров на складе"
        },
        "description": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_description",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 15:32:48.999118492 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "На данной странице электронного курса собраны все полезные материалы по процессу «Сборка товара». Изучи материалы и пройди небольшую проверку по пройденному материалу"
        },
        "label": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_label",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_order",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:30:09.154667608 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2"
        },
        "preview": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_preview",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "preview/Group 105.png"
        },
        "title": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_title",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "Электронный курс «Сборка товара»"
        },
        "to_build": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_type",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-28T13-14-02Z-941b6d_zip-scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 10:18:06.234025616 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 16:30:09.175300894 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Электронный курс «Сборка товара»",
      "type": "",
      "uid": "2023-11-28T13-14-02Z-941b6d"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c__datecreate",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "2023-11-28 13:29:01.519531158 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c__groups",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c__owner",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "WBProf Оператор"
        },
        "_title": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c__title",
          "editor": "",
          "rev": "2024-01-25 16:31:45.159181962 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Электронный курс «Сортировка товара»"
        },
        "access_admin": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_block",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "2023-11-13T10-51-21Z-624491",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "Процесс сортировки товаров на складе"
        },
        "description": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_description",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:07:23.186296266 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "На данной странице электронного курса собраны все полезные материалы по процессу «Сортировка товара». Изучи материалы и пройди небольшую проверку по пройденному материалу"
        },
        "label": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_label",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_order",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:31:45.106732384 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2"
        },
        "preview": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_preview",
          "editor": "",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "preview/Group 106.png"
        },
        "title": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_title",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "Электронный курс «Сортировка товара»"
        },
        "to_build": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_type",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-28T13-27-10Z-6a882c_zip-scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-11-29 07:42:54.178835087 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "/lms/gui/scorm/Sortirovka.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 16:31:45.159181962 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Электронный курс «Сортировка товара»",
      "type": "",
      "uid": "2023-11-28T13-27-10Z-6a882c"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d__datecreate",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "2023-11-28 12:52:17.860887224 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d__groups",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d__owner",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "WBProf Оператор"
        },
        "_title": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d__title",
          "editor": "",
          "rev": "2024-01-25 16:27:00.079194216 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Электронный курс по операции «Раскладка товара»"
        },
        "access_admin": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_block",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "2023-11-13T10-50-57Z-8f8663",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "Процесс раскладки товаров на мезонин"
        },
        "description": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_description",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "На данной странице электронного курса собраны все полезные материалы по процессу «Раскладка товара». Изучи материалы и пройди небольшую проверку по пройденному материалу"
        },
        "label": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_label",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_order",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:27:00.058559556 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2"
        },
        "preview": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_preview",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "preview/Group 103.png"
        },
        "start_url_scorm": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_start_url_scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "/scormcontent/index.html"
        },
        "title": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_title",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "Электронный курс по операции «Раскладка товара»"
        },
        "to_build": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_type",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-28T12-51-00Z-344d5d_zip-scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-08 08:20:54.034188385 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "scorm/raskladka_scorm.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 16:27:00.079194216 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Электронный курс по операции «Раскладка товара»",
      "type": "",
      "uid": "2023-11-28T12-51-00Z-344d5d"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20__datecreate",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "2023-11-23 09:13:25.0728241 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20__groups",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "_owner": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20__owner",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "WBProf Оператор"
        },
        "_title": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20__title",
          "editor": "",
          "rev": "2024-01-25 16:27:12.339436057 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Файл SCORM Электронный курс по процессу «Приемка товара»"
        },
        "access_admin": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_block",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "2023-11-13T10-50-45Z-f57031",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "Процесс приема товаров на складе"
        },
        "description": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_description",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "На данной странице электронного курса собраны все полезные материалы по процессу «Приемка товара». Изучи материалы и пройди небольшую проверку по пройденному материалу"
        },
        "label": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_label",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": ""
        },
        "order": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_order",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 16:27:12.314551431 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "2"
        },
        "preview": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_preview",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "preview/Group 120.png"
        },
        "start_url_scorm": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_start_url_scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "",
          "value": "/scormcontent/index.html"
        },
        "title": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_title",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2024-01-25 15:26:38.647295602 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "Электронный курс по процессу «Приемка товара»"
        },
        "to_build": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_type",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "tpl_lms_material_type_scorm",
          "status": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Файл SCORM"
        },
        "zip-scorm": {
          "Uuid": "2023-11-23T09-12-04Z-b58f20_zip-scorm",
          "editor": "5c29deb5-9cce-a153-331a-c14439ffa038",
          "rev": "2023-12-07 13:17:28.800078501 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "scorm/priemka_scorm.zip"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2024-01-25 16:27:12.339436057 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "Файл SCORM Электронный курс по процессу «Приемка товара»",
      "type": "",
      "uid": "2023-11-23T09-12-04Z-b58f20"
    },
    {
      "attributes": {
        "_datecreate": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0__datecreate",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "",
          "value": "2023-11-08 19:38:27.654021759 +0000 UTC"
        },
        "_groups": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0__groups",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "2021-06-21T10-16-22z03-00-1d3e17,2020-05-23T06-19-07Z-1d6d78",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "2020-05-23T06-14-40Z-e312fc",
          "value": "Администраторы, Разработчики"
        },
        "_owner": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0__owner",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "679ee02b-b537-4aa3-b91d-589d17826ba1",
          "value": "Иван Ловецкий"
        },
        "_title": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0__title",
          "editor": "",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "тест материал Видео 12312312"
        },
        "access_admin": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_delete": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_read": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "access_write": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать роль -"
        },
        "block": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_block",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "2023-11-08T19-36-54Z-0132c0",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "2023-09-05T17-36-16Z-5ce7ce",
          "value": "12312312"
        },
        "description": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_description",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "",
          "value": ""
        },
        "label": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_label",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "",
          "value": ""
        },
        "preview": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_preview",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "/lms/gui/preview/test_longrid.html"
        },
        "title": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_title",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "",
          "value": "тест материал"
        },
        "to_build": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": ""
        },
        "to_update": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "",
          "rev": "",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "- выбрать сервер -"
        },
        "type": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_type",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "tpl_lms_material_type_video",
          "status": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "tpls": "2023-09-06T06-27-01Z-a4186a",
          "value": "Видео"
        },
        "url_video": {
          "Uuid": "2023-11-08T19-38-09Z-d8c2f0_url_video",
          "editor": "7929d995-3206-d127-fa4a-1e6950eaaa22",
          "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
          "src": "",
          "status": "",
          "tpls": "",
          "value": "123.mp4"
        }
      },
      "copies": "",
      "id": "",
      "parent": "2023-09-05T17-29-49Z-ddc679",
      "rev": "2023-11-13 10:44:58.490754046 +0000 UTC",
      "source": "2023-09-05T17-29-49Z-ddc679",
      "title": "тест материал Видео 12312312",
      "type": "",
      "uid": "2023-11-08T19-38-09Z-d8c2f0"
    }
  ],
  "status": {
    "description": "",
    "status": 200,
    "code": "",
    "error": ""
  },
  "metrics": {
    "result_size": 160,
    "result_count": 160,
    "result_offset": 0,
    "result_limit": 250,
    "result_page": 1,
    "time_execution": "891.101305ms",
    "time_query": "map[CalcUIDs:10.386576ms GO:878.639486ms SLP:2.063398ms]",
    "page_last": 1,
    "page_current": 1,
    "page_list": [
      1
    ],
    "page_from": 0,
    "page_to": 250
  }
}`

	var obj models.ResponseData
	json.Unmarshal([]byte(in), &obj)

	for _, v := range obj.Data {
		println(v.Title)
	}

	fmt.Println("-----------------")

	NewFuncMap(nil, nil, "")
	res, err := Funcs.sortbyfield(obj, "", "rev", true)
	if err != nil {
		t.Errorf("Should not produce an error, err: %s", err)
	}

	b1, _ := json.Marshal(res)
	var obj2 models.ResponseData
	err = json.Unmarshal(b1, &obj2)
	if err != nil {
		t.Errorf("Unmarshal, err: %s", err)
	}

	for _, v := range obj2.Data {
		println(v.Attr("", "rev"))
	}
}

func Test_funcMap_convert(t1 *testing.T) {
	// Создание фейковых данных для тестирования
	contentUTF8 := []byte("Пример текста на русском языке")
	contentUTF8BOM := append([]byte{0xEF, 0xBB, 0xBF}, contentUTF8...)
	contentWindows1251 := []byte{207, 240, 232, 236, 229, 240, 32, 242, 229, 234, 241, 242, 224, 32, 237, 224, 32, 240, 243, 241, 241, 234, 238, 236, 32, 255, 231, 251, 234, 229}

	type args struct {
		content        []byte
		targetEncoding string
	}
	tests := []struct {
		name            string
		args            args
		wantEncodedData []byte
	}{
		// UTF-8
		{
			name: "UTF-8 to UTF-8",
			args: args{
				content:        contentUTF8,
				targetEncoding: "UTF-8",
			},
			wantEncodedData: contentUTF8,
		},
		{
			name: "UTF-8 to UTF-8 BOM",
			args: args{
				content:        contentUTF8,
				targetEncoding: "UTF-8 BOM",
			},
			wantEncodedData: contentUTF8BOM,
		},
		{
			name: "UTF-8 to windows-1251",
			args: args{
				content:        contentUTF8,
				targetEncoding: "windows-1251",
			},
			wantEncodedData: contentWindows1251,
		},

		// UTF-8 BOM
		{
			name: "UTF-8 BOM to UTF-8",
			args: args{
				content:        contentUTF8BOM,
				targetEncoding: "UTF-8",
			},
			wantEncodedData: contentUTF8,
		},
		{
			name: "UTF-8 BOM to UTF-8 BOM",
			args: args{
				content:        contentUTF8BOM,
				targetEncoding: "UTF-8 BOM",
			},
			wantEncodedData: contentUTF8BOM,
		},
		{
			name: "UTF-8 BOM to windows-1251",
			args: args{
				content:        contentUTF8BOM,
				targetEncoding: "windows-1251",
			},
			wantEncodedData: contentWindows1251,
		},
		// windows-1251
		{
			name: "windows-1251 to UTF-8",
			args: args{
				content:        contentWindows1251,
				targetEncoding: "UTF-8",
			},
			wantEncodedData: contentUTF8,
		},
		{
			name: "windows-1251 to UTF-8 BOM",
			args: args{
				content:        contentWindows1251,
				targetEncoding: "UTF-8 BOM",
			},
			wantEncodedData: contentUTF8BOM,
		},
		{
			name: "windows-1251 to windows-1251",
			args: args{
				content:        contentWindows1251,
				targetEncoding: "windows-1251",
			},
			wantEncodedData: contentWindows1251,
		},

		{
			name: "err",
			args: args{
				content:        []byte(""),
				targetEncoding: "windows-1251",
			},
			wantEncodedData: []byte(""),
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &funcMap{}
			if gotEncodedData := t.convert(tt.args.content, tt.args.targetEncoding); !reflect.DeepEqual(gotEncodedData, tt.wantEncodedData) {
				t1.Errorf("convert() = %v, want %v", gotEncodedData, tt.wantEncodedData)
			}
		})
	}
}

func Test_decodebase64(t *testing.T) {
	in := "user1:passw0rd"

	NewFuncMap(nil, nil, "")
	res := Funcs.decodebase64(in)

	fmt.Println(res)
	res = Funcs.encodebase64(res)

	fmt.Println(res)
}

func Test_loggert(t *testing.T) {
	cfg := config

	cfg.CbMaxRequestsLogbox = 3
	cfg.CbTimeoutLogbox = 5 * time.Second
	cfg.CbIntervalLogbox = 5 * time.Second
	cfg.LogboxEndpoint = "http://127.0.0.1:8999"
	cfg.CbMaxRequestsLogbox = 3
	cfg.CbTimeoutLogbox = 5 * time.Second
	cfg.CbIntervalLogbox = 5 * time.Second

	err := logger.SetupDefaultLogboxLogger("app/client", logger.LogboxConfig{
		Endpoint:       cfg.LogboxEndpoint,
		AccessKeyID:    cfg.LogboxAccessKeyId,
		SecretKey:      cfg.LogboxSecretKey,
		RequestTimeout: cfg.LogboxRequestTimeout,
		CbMaxRequests:  cfg.CbMaxRequestsLogbox,
		CbTimeout:      cfg.CbTimeoutLogbox,
		CbInterval:     cfg.CbIntervalLogbox,
	}, map[string]string{
		logger.ServiceIDKey:   lib.Hash(lib.UUID()),
		logger.ConfigIDKey:    "app",
		logger.ServiceTypeKey: "app",
	})

	fmt.Println(err)

	NewFuncMap(nil, nil, "")
	res := Funcs.logger("info", "test", "key1", "value1", "key2", "value2")

	fmt.Println(res)
}
