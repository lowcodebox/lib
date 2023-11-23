package service

import (
	"context"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// Files ...
func (s *service) Files(ctx context.Context, in model.ServiceFilesIn) (out model.ServiceFilesOut, err error) {
	if in.Action == model.FilesActionLoad {
		out, err = loadFileOnly(ctx, in)
	}

	return out, err
}

func loadFileOnly(ctx context.Context, in model.ServiceFilesIn) (out model.ServiceFilesOut, err error) {
	//var path = ""

	//// добавляем переданный путь
	//getPath := r.FormValue("path")
	//getName := r.FormValue("name")
	//
	//urlPath := strings.Split(getPath, sep)
	//for _, v := range urlPath {
	//	path = filepath.Join(path, v)
	//}
	//
	//// полный путь к файлу
	//thisFilePath := path + sep + getName
	//objResp.Status.Description = thisFilePath
	//
	//err = h.vfs.Write(thisFilePath, []byte{})
	//if err != nil {
	//	logger.Error(h.ctx, "error vfs Write", zap.Error(err))
	//	objResp.Status.Error = fmt.Sprintf("error vfs Write, err: %s", err)
	//}
	//
	//// формируем ответ
	//outByte, err := json.Marshal(objResp)
	//if err != nil {
	//	out = fmt.Sprint(err)
	//	return
	//}
	//
	//out = string(outByte)

	return
}
