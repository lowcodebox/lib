package app_lib

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strings"
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

type Manifest struct {
	Resources     Resources     `xml:"resources"`
	Organizations Organizations `xml:"organizations"`
}

func (t *FuncMapImpl) parsescorm(zipFilename string, destPath string) (index string) {
	folder := t.unzip(zipFilename, destPath)

	files, err := t.vfs.List(context.Background(), folder, math.MaxInt32)
	if err != nil {
		return fmt.Sprintf("error parsescorm vfs.List, err: %s", err)
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
				return fmt.Sprintf("error parsescorm file.Open, err: %s", err)
			}
			defer rc.Close()

			d, err := io.ReadAll(rc)
			if err != nil {
				return fmt.Sprintf("error parsecorm io.ReadAll, err: %s", err)
			}

			var manifest Manifest
			err = xml.Unmarshal(d, &manifest)
			if err != nil {
				return fmt.Sprintf("error parsescorm xml.Unmarshal, err: %s", err)
			}

			resourceId := manifest.Organizations.Organization[0].Item.Identifierref

			for _, resource := range manifest.Resources.Resource {
				if resource.Identifier == resourceId {
					index = resource.Href
				}
			}

			break
		}
	}

	//до этого нашли название файла, а теперь ищем полный путь до него
	for _, file := range files {
		if strings.Contains(file.Name(), index) {
			index = file.Name()
			return index
		}
	}

	return "error index file not found"
}
