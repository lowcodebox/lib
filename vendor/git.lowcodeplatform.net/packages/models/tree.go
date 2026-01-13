package models

type DataTree struct {
	Data
	Sub  []string    `json:"sub"`
	Incl []*DataTree `json:"incl"`
}

type DataTreeOut struct {
	Data
	Sub  []string   `json:"sub"`
	Incl []DataTree `json:"incl"`
}

////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Attr возвращаем необходимый значение атрибута для объекта если он есть, инае пусто
// а также из заголовка объекта
func (p *DataTree) Attr(name, element string) (result string, found bool) {

	if _, found := p.Attributes[name]; found {

		// фикс для тех объектов, на которых добавлено скрытое поле Uid
		if name == "uid" {
			return p.Uid, true
		}

		switch element {
		case "src":
			return p.Attributes[name].Src, true
		case "value":
			return p.Attributes[name].Value, true
		case "tpls":
			return p.Attributes[name].Tpls, true
		case "rev":
			return p.Attributes[name].Rev, true
		case "status":
			return p.Attributes[name].Status, true
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		case "copies":
			return p.Copies, true
		}
	} else {
		switch element {
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "rev":
			return p.Rev, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		case "copies":
			return p.Copies, true
		}
	}
	return "", false
}

// метод типа Items (перемещаем структуры в карте, исходя из заявленной вложенности элементов)
// (переделать дубль фукнции)
func (p *DataTree) ScanSub(maps *map[string]*DataTree) {
	if p.Sub != nil && len(p.Sub) != 0 {
		for _, c := range p.Sub {
			gg := *maps
			fromP := gg[c]
			if fromP != nil {
				copyPolygon := *fromP
				p.Incl = append(p.Incl, &copyPolygon)
				delete(*maps, c)
				copyPolygon.ScanSub(maps)
			}
		}
	}
}
