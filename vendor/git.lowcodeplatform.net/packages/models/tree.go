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
