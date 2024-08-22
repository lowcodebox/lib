package app_lib

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

const ImsManifest = "imsmanifest.xml"

type ItemXML struct {
	Identifierref string `xml:"identifierref,attr"`
}

type Organization struct {
	Item       ItemXML `xml:"item"`
	Identifier string  `xml:"identifier,attr"`
}

type Organizations struct {
	Organization []Organization `xml:"organization"`
}

type Resource struct {
	Identifier string `xml:"identifier,attr"`
	Href       string `xml:"href,attr"`
}

type Resources struct {
	Resource []Resource `xml:"resource"`
}

type Metadata struct {
	Schema        string `xml:"schema"`
	Schemaversion string `xml:"schemaversion"`
}

type Manifest struct {
	Resources     Resources     `xml:"resources"`
	Organizations Organizations `xml:"organizations"`
	Metadata      Metadata      `xml:"metadata"`
}

type ScormRes struct {
	Version string
	Index   string
	Error   string
}

func (t *FuncImpl) parsescorm(zipFilename string, destPath string) (sr ScormRes) {
	folder := t.unzip(zipFilename, destPath)

	files, err := t.vfs.List(context.Background(), folder, math.MaxInt32)
	if err != nil {
		sr.Error = fmt.Sprintf("error parsescorm vfs.List, err: %s", err)
		return
	}

	for _, file := range files {

		// в макоси создаются файлы с припиской __MACOSX/, например:
		// somedirectory/somefile.extension
		// __MACOSX/somedirectory/._somefile.extension
		// где нужен только первый файл
		if strings.HasPrefix(file.Name(), "__MACOSX/") {
			continue
		}

		//в файле imsmanifest.xml содержится инфа о стартовом html
		if strings.Contains(file.Name(), ImsManifest) {
			rc, err := file.Open()
			if err != nil {
				sr.Error = fmt.Sprintf("error parsescorm file.Open, err: %s", err)
				return
			}
			defer rc.Close()

			d, err := io.ReadAll(rc)
			if err != nil {
				sr.Error = fmt.Sprintf("error parsecorm io.ReadAll, err: %s", err)
				return
			}

			var manifest Manifest
			err = xml.Unmarshal(d, &manifest)
			if err != nil {
				sr.Error = fmt.Sprintf("error parsescorm xml.Unmarshal, err: %s", err)
				return
			}

			sr.Version = manifest.Metadata.Schemaversion
			resourceId := manifest.Organizations.Organization[0].Item.Identifierref

			for _, resource := range manifest.Resources.Resource {
				if resource.Identifier == resourceId {
					indexSlice := strings.Split(resource.Href, "?")
					sr.Index = indexSlice[0]
				}
			}

			break
		}
	}

	//до этого нашли название файла, а теперь ищем полный путь до него
	for _, file := range files {
		if strings.Contains(file.Name(), sr.Index) {
			sr.Index = file.Name()
			return
		}
	}
	sr.Error = "error index file not found"
	return
}

func (t *FuncImpl) RecursiveChildren(parentUid string, relationField string, recursiveLevel int) (result models.ResponseData) {
	var findChildren func(parent models.Data, relationField string, recursiveCalls int) error

	findChildren = func(parent models.Data, relationField string, recursiveCalls int) error {
		if (recursiveCalls > recursiveLevel) && (recursiveLevel != 0) {
			return nil
		}

		//Делаем тело запроса на получение потомков родителя
		searchParams := map[string]string{
			"tpls":         parent.Source,
			"limit":        "100",
			"filter_src":   parent.Uid,
			"filter_field": relationField,
			"short":        "false",
		}

		//fmt.Printf("searchParams: %+v\n", searchParams)

		p, err := json.Marshal(searchParams)
		if err != nil {
			return err
		}

		childrenStr, err := t.api.Search(context.Background(), "apiSearch", http.MethodPost, string(p))
		if err != nil {
			return err
		}

		var children models.ResponseData
		err = json.Unmarshal([]byte(childrenStr), &children)
		if err != nil {
			return err
		}

		//Если нет детей, то дошли до низа рекурсии, можем идти наверх
		if len(children.Data) == 0 {
			return nil
		}

		for _, child := range children.Data {
			//Если потомок не такого же шаблона то не добавляем его
			if child.Source != parent.Source {
				continue
			}
			//Добавляем потомка
			result.Data = append(result.Data, child)

			//Ищем потомков потомка
			err := findChildren(child, relationField, recursiveCalls+1)
			if err != nil {
				return err
			}

		}

		return nil
	}

	//Ищем объект главного предка
	parentRD, err := t.api.ObjGet(context.Background(), parentUid)
	if err != nil {
		result.Status.Error = err
		return
	}

	if len(parentRD.Data) == 0 {
		result.Status.Error = err
		return
	}

	//Ищем потомков родителя
	if err := findChildren(parentRD.Data[0], relationField, 1); err != nil {
		result.Status.Error = err
		return
	}

	return
}
