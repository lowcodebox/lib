package app_lib

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
