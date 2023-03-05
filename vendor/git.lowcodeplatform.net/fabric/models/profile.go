package models

type ProfileData struct {
	Revision       string `json:"revision"` // ревизия текущей сессии (если сессия обновляется (меняется профиль) - ID-сессии остается, но ревизия меняется
	Hash           string `json:"hash"`
	Email          string `json:"email"`
	Uid            string `json:"uid"`
	ObjUid         string `json:"obj_uid"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Photo          string `json:"photo"`
	Age            string `json:"age"`
	City           string `json:"city"`
	Country        string `json:"country"`
	Oauth_identity string `json:"oauth_identity"`
	Status         string `json:"status"` // - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw            []Data `json:"raw"`    // объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables         []Data `json:"tables"`
	Roles          []Data
	Homepage       string   `json:"homepage"`
	Maket          string   `json:"maket"`
	UpdateFlag     bool     `json:"update_flag"`
	UpdateData     []Data   `json:"update_data"`
	CurrentRole    Data     `json:"current_role"`
	Profiles       []Data   `json:"profiles"`
	CurrentProfile Data     `json:"current_profile"`
	Navigator      []*Items `json:"navigator"`

	Groups             string
	GroupsValue        string
	GroupsDefaultSrc   string
	GroupsDefaultValue string

	ButtonsNavTop []Data
	CountLicense  int
	BaseMode      map[string]string

	// TODO проверить где используется и выпилить
	RolesOld   map[string]string `json:"roles"` //deplicated
	First_name string            //deplicated
	Last_name  string            //deplicated

}

type Items struct {
	Title        string   `json:"title"`
	ExtentedLink string   `json:"extentedLink"`
	Uid          string   `json:"uid"`
	Source       string   `json:"source"`
	Icon         string   `json:"icon"`
	Leader       string   `json:"leader"`
	Order        string   `json:"order"`
	Type         string   `json:"type"`
	Preview      string   `json:"preview"`
	Url          string   `json:"url"`
	Sub          []string `json:"sub"`
	Incl         []*Items `json:"incl"`
	Class        string   `json:"class"`
	FinderMode   string   `json:"finder_mode"`
}

// ScanSub метод типа Items (перемещаем структуры в карте, исходя из заявленной вложенности элементов)
// (переделать дубль фукнции)
func (p *Items) ScanSub(maps *map[string]*Items) {
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
