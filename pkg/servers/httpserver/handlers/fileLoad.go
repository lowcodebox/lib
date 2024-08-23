package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/segmentio/ksuid"
)

const MByte = 1000000

func BodyToResponse(w http.ResponseWriter, objResp *models.Response, flagCKEditor bool) {
	// NO TESTED
	if flagCKEditor { // для CKEditor уже отправили ответ в другом формате
		return
	}

	// формируем ответ
	out, err := json.Marshal(*objResp)
	if err != nil {
		objResp.Status.Error = err
	}

	//w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//w.Header().Set("Accept", "application/json")
	w.WriteHeader(objResp.Status.Status)
	fmt.Fprintln(w, string(out))

}

func parseForm(r *http.Request) (string, string, string, string, string, int64, error) {
	fileField := "uploadfile"
	if r.FormValue("CKEditor") != "" {
		fileField = "upload"
	}
	getPath := r.FormValue("path")
	getField := r.FormValue("field")
	mode := r.FormValue("mode")
	objuid := r.FormValue("objuid")
	contentLength := r.ContentLength

	return fileField, getPath, getField, mode, objuid, contentLength, nil
}

func CKEditorHandler(w http.ResponseWriter, r *http.Request, thisFilePath string) {
	num := r.FormValue("CKEditorFuncNum")
	path := "/upload" + thisFilePath
	outScript := `
		<script type="text/javascript">
			window.parent.CKEDITOR.tools.callFunction('` + num + `', '` + path + `','');
		</script>
	`
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Write([]byte(path))
	w.Write([]byte(outScript))
}

func (h *handlers) objLoad(r *http.Request, objuid, getField, mode string, file io.Reader, handler *multipart.FileHeader, contentLength int64, thisFilePath *string, objResp *models.Response) error {
	var sliceFiles []string

	obj, err := h.api.ObjGet(r.Context(), objuid)
	if err != nil {
		return err
	}

	for _, v := range obj.Data {
		source := v.Source
		elements, err := h.api.Element(h.ctx, "elements", source)
		if err != nil || len(elements.Data) == 0 {
			return errors.New("error getting elements")
		}

		foundFileField := false
		for _, el := range elements.Data {
			id := el.Id
			if id == getField {
				foundFileField = true
				maxSizeString, f := el.Attr("max_size", "value")
				if f {
					maxSize, err := strconv.ParseFloat(maxSizeString, 64)
					if err == nil && contentLength > int64(maxSize*MByte) {
						return errors.New("error too big file")
					}
				}
				break
			}
		}
		if !(foundFileField) && (contentLength > 50*MByte) {
			return errors.New("error too big file")
		}
	}

	if mode == "multi" {
		for _, v := range obj.Data {
			path, found := v.Attr(getField, "value")
			if !found {
				return errors.New("missing field")
			}
			sliceFiles = strings.Split(path, ",")
		}
	}

	sliceFiles = append(sliceFiles, *thisFilePath)
	*thisFilePath = strings.Join(sliceFiles, ",")

	res := make([]byte, handler.Size)
	_, err = io.ReadFull(file, res)
	if err != nil {
		return err
	}

	err = h.vfs.Write(h.ctx, *thisFilePath, res)
	if err != nil {
		return err
	}

	_, err = h.api.ObjAttrUpdate(h.ctx, objuid, getField, *thisFilePath, "", "")
	if err != nil {
		return err
	}

	objResp.Status.Description = *thisFilePath
	return nil
}

func (h *handlers) noObjLoad(file io.Reader, handler *multipart.FileHeader, contentLength int64, thisFilePath *string, objResp *models.Response) error {
	if contentLength > 10*MByte {
		return errors.New("error too big file")
	}

	res := make([]byte, handler.Size)
	_, err := io.ReadFull(file, res)
	if err != nil {
		return err
	}

	err = h.vfs.Write(h.ctx, *thisFilePath, res)
	if err != nil {
		return err
	}

	objResp.Status.Description = *thisFilePath
	return nil
}

func (h *handlers) FileLoad(w http.ResponseWriter, r *http.Request) {
	var objResp models.Response

	objResp.Status.Status = 200
	defer BodyToResponse(w, &objResp, r.FormValue("CKEditor") != "")

	fileField, getPath, getField, mode, objuid, contentLength, err := parseForm(r)
	if err != nil {
		objResp.Status.Error = err
		return
	}

	file, handler, err := r.FormFile(fileField)
	if err != nil {
		objResp.Status.Error = err
		return
	}
	defer file.Close()

	path := filepath.Join(strings.Split(getPath, "/")...)

	thisFilePath := path + sep + handler.Filename
	data, _, _ := h.vfs.Read(h.ctx, thisFilePath)
	if len(data) != 0 {
		thisFilePath = path + sep + ksuid.New().String() + "_" + handler.Filename
	}

	if objuid != "" {
		err = h.objLoad(r, objuid, getField, mode, file, handler, contentLength, &thisFilePath, &objResp)
		if err != nil {
			objResp.Status.Error = err
			return
		}
	} else {
		err = h.noObjLoad(file, handler, contentLength, &thisFilePath, &objResp)
		if err != nil {
			objResp.Status.Error = err
			return
		}
	}

	if r.FormValue("CKEditor") != "" {
		CKEditorHandler(w, r, thisFilePath)
		return
	}
}
