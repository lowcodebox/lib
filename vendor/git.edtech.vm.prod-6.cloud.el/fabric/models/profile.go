package models

type ProfileData struct {
	Revision       string `json:"revision,omitempty"` // ревизия текущей сессии (если сессия обновляется (меняется профиль) - ID-сессии остается, но ревизия меняется
	Hash           string `json:"hash,omitempty"`
	Email          string `json:"email,omitempty"`
	Uid            string `json:"uid,omitempty"`
	ObjUid         string `json:"obj_uid,omitempty"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	FatherName     string `json:"father_name,omitempty"`
	PatronName     string `json:"patron_name,omitempty"`
	Photo          string `json:"photo,omitempty"`
	Age            string `json:"age,omitempty"`
	City           string `json:"city,omitempty"`
	Country        string `json:"country,omitempty"`
	Oauth_identity string `json:"oauth_identity,omitempty"`
	Status         string `json:"status,omitempty"` // - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw            []Data `json:"raw,omitempty"`    // объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables         []Data `json:"tables,omitempty"`
	Roles          []Data // разремить после запуска новой версии
	//Roles          map[string]string `json:"roles"` // deprecated
	Homepage       string   `json:"homepage,omitempty"`
	Maket          string   `json:"maket,omitempty"`
	UpdateFlag     bool     `json:"update_flag"`
	UpdateData     []Data   `json:"update_data,omitempty"`
	CurrentRole    Data     `json:"current_role,omitempty"`
	Profiles       []Data   `json:"profiles,omitempty"`
	CurrentProfile Data     `json:"current_profile,omitempty"`
	Navigator      []*Items `json:"navigator,omitempty"`

	Groups             string
	GroupsValue        string
	GroupsDefaultSrc   string
	GroupsDefaultValue string

	ButtonsNavTop []Data
	CountLicense  int
	BaseMode      map[string]string

	// TODO проверить где используется и выпилить
	RolesOld   map[string]string `json:"roles_old,omitempty"` //deplicated
	First_name string            //deplicated
	Last_name  string            //deplicated

	Identity        string `json:"identity,omitempty"`
	PrimaryIdentity string `json:"primary_identity,omitempty"`
	Phone           string `json:"phone,omitempty"`

	EmployeeId string `json:"employee_id,omitempty"`
	UserId     string `json:"user_id,omitempty"`
	OfficeId   string `json:"office_id,omitempty"`
	ShardId    string `json:"shard_id,omitempty"`

	// IAM будет искать по этим RoleIDs роли, у которых сходится атрибут role_id, и создавать профили с этими ролями
	RoleIDs string `json:"role_ids"`

	BirthDate string `json:"birth_date"`
	Login     string `json:"login"`
}

type Items struct {
	Title        string   `json:"title,omitempty"`
	ExtentedLink string   `json:"extentedLink,omitempty"`
	Uid          string   `json:"uid,omitempty"`
	Source       string   `json:"source,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	Leader       string   `json:"leader,omitempty"`
	LeaderValue  string   `json:"leader_value,omitempty"`
	Order        string   `json:"order,omitempty"`
	Type         string   `json:"type,omitempty"`
	Preview      string   `json:"preview,omitempty"`
	Url          string   `json:"url,omitempty"`
	Sub          []string `json:"sub,omitempty"`
	Incl         []*Items `json:"incl,omitempty"`
	Class        string   `json:"class,omitempty"`
	FinderMode   string   `json:"finder_mode,omitempty"`
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
