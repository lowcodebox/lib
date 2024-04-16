package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

var sep = string(filepath.Separator)

// Files operation from files
// @Router /api/v1/files [get/post/put/delete]
func (h *handlers) Files(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Files] Error response execution",
				zap.String("url", r.RequestURI),
				zap.Error(err))
		}
	}()

	in, er := filesDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesDecodeRequest")
		return
	}

	serviceResult, er := h.service.Files(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec service.Files")
		return
	}

	response, er := filesEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesEncodeResponse")
		return
	}

	err = filesTransportResponse(r.Context(), w, response)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesTransportResponse")
		return
	}

	return
}

func filesDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceFilesIn, err error) {
	if r.Method == http.MethodPost {
		in.Action = model.FilesActionLoad
	}
	return in, err
}

func filesEncodeResponse(ctx context.Context, serviceResult model.ServiceFilesOut) (response model.ServiceFilesOut, err error) {
	return serviceResult, err
}

func filesTransportResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	d, err := json.Marshal(response)
	if err != nil {
		logger.Error(ctx, "[Files] (filesTransportResponse) error ParseForm", zap.Error(err))
	}

	_, err = w.Write(d)
	if err != nil {
		logger.Error(ctx, "[Files] (filesTransportResponse) error Write", zap.Error(err))
	}

	return err
}

func (h *handlers) FileLoad(w http.ResponseWriter, r *http.Request) {
	var objResp models.Response
	var path = ""
	objResp.Status.Status = 200
	flagCKEditor := false
	fileField := "uploadfile"

	defer func() {
		// NO TESTED
		if flagCKEditor { // для CKEditor уже отправили ответ в другом формате
			return
		}

		// формируем ответ
		out, err := json.Marshal(objResp)
		if err != nil {
			objResp.Status.Error = err
			log.Error(err)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Accept", "application/json")
		w.WriteHeader(objResp.Status.Status)
		w.Write(out)
	}()

	if r.FormValue("CKEditor") != "" {
		fileField = "upload"
		flagCKEditor = true
	}

	file, handler, err := r.FormFile(fileField)
	if err != nil {
		objResp.Status.Error = err
		return
	}
	defer file.Close()

	// добавляем переданный путь
	getPath := r.FormValue("path")
	// название поля для загрузки
	getField := r.FormValue("field")
	// название поля для загрузки
	mode := r.FormValue("mode")

	r.Form.Set("os", "linux")
	getPath = h.app.DogParse(getPath, r, nil, nil)

	// для обработки возможных встроенных @-фукнций если передан objuid - берем этот объект
	objuid := r.FormValue("objuid")
	if objuid != "" {
		var objProduct models.ResponseData
		h.app.Curl("GET", "_objs/"+objuid, "", &objProduct, nil)
		getPath = h.app.DogParse(getPath, r, &objProduct.Data, nil)
	}
	urlPath := strings.Split(getPath, "/")

	for _, v := range urlPath {
		path = filepath.Join(path, v)
	}

	// все операции с файлами происходят через VFS
	// полный путь к файлу
	thisFilePath := path + sep + handler.Filename
	objResp.Status.Description = thisFilePath

	res := make([]byte, handler.Size)
	_, err = io.ReadFull(file, res)
	if err != nil {
		objResp.Status.Error = err
		return
	}

	err = h.vfs.Write(h.ctx, thisFilePath, res)
	if err != nil {
		log.Error(err)
		objResp.Status.Error = err
		return
	}

	// ответ для CKEditor-a
	if flagCKEditor {
		num := r.FormValue("CKEditorFuncNum")
		path = h.app.ConfigGet("ClientPath") + "/upload" + thisFilePath
		outScript := `
 			<script type="text/javascript">
				window.parent.CKEDITOR.tools.callFunction('` + num + `', '` + path + `','');
			</script>
		`
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Write([]byte(path))
		w.Write([]byte(outScript))

		return
	}

	// обновляем значение в поле данных загруженных данных объекта
	// если multi - сохраняем как список через ,
	// иначе подменяем значение старого пути
	var sliceFiles = []string{}
	if mode == "multi" {
		obj, err := h.api.ObjGet(r.Context(), objuid)
		if err != nil {
			objResp.Status.Error = err
		}
		for _, v := range obj.Data {
			path, found := v.Attr(getField, "value")
			if !found {
				objResp.Status.Error = err
				return
			}
			sliceFiles = strings.Split(path, ",")
		}
	}

	sliceFiles = append(sliceFiles, thisFilePath)
	thisFilePath = strings.Join(sliceFiles, ",")

	_, err = h.api.ObjAttrUpdate(h.ctx, objuid, getField, thisFilePath, "", "")
	if err != nil {
		objResp.Status.Error = err
	}

	return
}
