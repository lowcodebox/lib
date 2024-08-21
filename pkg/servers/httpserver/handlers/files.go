package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/labstack/gommon/log"
	"github.com/segmentio/ksuid"
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

	contentLength := 0.0

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
	logger.Info(h.ctx, "fileload formfile",
		zap.String("filename", handler.Filename),
		zap.Int64("filesize", handler.Size),
	)

	if err != nil {
		objResp.Status.Error = err
		logger.Error(h.ctx, "fileload formfile error",
			zap.String("filename", handler.Filename),
			zap.Int64("filesize", handler.Size),
		)
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

	objuid := r.FormValue("objuid")

	urlPath := strings.Split(getPath, "/")

	for _, v := range urlPath {
		path = filepath.Join(path, v)
	}

	contentLengthString := r.Header.Get("Content-Length")
	if contentLengthString != "" {
		contentLength, err = strconv.ParseFloat(contentLengthString, 64)
		if err != nil {
			objResp.Status.Error = err
			return
		}
	}

	// все операции с файлами происходят через VFS
	// полный путь к файлу
	thisFilePath := path + sep + handler.Filename
	// вычитываем файл из хранилища (чтобы проверить если уже есть такой, чтобы не затереть старый)
	data, _, _ := h.vfs.Read(h.ctx, thisFilePath)
	if len(data) != 0 {
		thisFilePath = path + sep + ksuid.New().String() + "_" + handler.Filename
	}
	objResp.Status.Description = thisFilePath

	if objuid != "" {
		obj, err := h.api.ObjGet(r.Context(), objuid)
		if err != nil {
			objResp.Status.Error = err
			return
		}

		if len(obj.Data) == 0 {
			objResp.Status.Error = errors.New("error object not found")
		}

		for _, v := range obj.Data {
			source := v.Source

			elements, err := h.api.Element(h.ctx, "elements", source)
			if err != nil {
				objResp.Status.Error = errors.New("error getting elements")
				return
			}

			for _, el := range elements.Data {
				id := el.Id
				if id == getField {
					maxSizeString, f := el.Attr("max_size", "value")
					if !f {
						//Не нашелся аттрибут max_size значит не задали макс размер, значит по умолчанию пропускаем все размеры
						break
					}
					maxSize, err := strconv.ParseFloat(maxSizeString, 64)
					if err != nil {
						//Аттрибут max_size нашелся, но его стерли поэтому не может запарсить, пропускаем как размер по умолчанию
						break
					}
					if (contentLength != 0.0) && (contentLength > maxSize*1000000) {
						objResp.Status.Error = errors.New("error too big file")
						return
					}
				}
			}
		}

		res := make([]byte, handler.Size)
		_, err = io.ReadFull(file, res)

		if err != nil {
			logger.Error(h.ctx, "fileload readfull err", zap.Error(err))
			objResp.Status.Error = err
			return
		}

		err = h.vfs.Write(h.ctx, thisFilePath, res)
		if err != nil {
			logger.Error(h.ctx, "fileload write err", zap.Error(err))
			objResp.Status.Error = err
			return
		}

		// обновляем значение в поле данных загруженных данных объекта
		// если multi - сохраняем как список через ,
		// иначе подменяем значение старого пути
		var sliceFiles = []string{}
		if mode == "multi" {
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
	} else {

		//Если нет шаблона, огран по умолчанию 10 мбайт
		if contentLength > 10000000 {
			objResp.Status.Error = errors.New("error too big file")
			return
		}

		res := make([]byte, handler.Size)
		_, err = io.ReadFull(file, res)

		if err != nil {
			logger.Error(h.ctx, "fileload readfull err", zap.Error(err))
			objResp.Status.Error = err
			return
		}

		err = h.vfs.Write(h.ctx, thisFilePath, res)
		if err != nil {
			logger.Error(h.ctx, "fileload write err", zap.Error(err))
			objResp.Status.Error = err
			return
		}
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

	return
}
