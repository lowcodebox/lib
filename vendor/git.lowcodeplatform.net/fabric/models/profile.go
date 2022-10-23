package models

type ProfileData struct {
	Revision		string `json:"revision"`	// ревизия текущей сессии (если сессия обновляется (меняется профиль) - ID-сессии остается, но ревизия меняется
	Hash       		string `json:"hash"`
	Email       	string `json:"email"`
	Uid         	string `json:"uid"`
	ObjUid			string `json:"obj_uid"`
	FirstName  		string `json:"first_name"`
	LastName   		string `json:"last_name"`
	Photo       	string `json:"photo"`
	Age       		string `json:"age"`
	City        	string `json:"city"`
	Country     	string `json:"country"`
	Oauth_identity	string `json:"oauth_identity"`
	Status 			string `json:"status"` 	// - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw	       		[]Data `json:"raw"`	// объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables      	[]Data `json:"tables"`
	Roles       	map[string]string `json:"roles"`
	Homepage		string `json:"homepage"`
	Maket			string `json:"maket"`
	UpdateFlag 		bool `json:"update_flag"`
	UpdateData 		[]Data `json:"update_data"`
	CurrentRole 	Data `json:"current_role"`
	CurrentProfile 	Data `json:"current_profile"`
	Navigator   	[]*Items `json:"navigator"`
}


type Items struct {
	Title  			string   	`json:"title"`
	ExtentedLink 	string 		`json:"extentedLink"`
	Uid    			string   	`json:"uid"`
	Source 			string   	`json:"source"`
	Icon   			string   	`json:"icon"`
	Leader 			string   	`json:"leader"`
	Order  			string   	`json:"order"`
	Type   			string   	`json:"type"`
	Preview			string   	`json:"preview"`
	Url    			string   	`json:"url"`
	Sub    			[]string 	`json:"sub"`
	Incl   			[]*Items 	`json:"incl"`
	Class 			string 		`json:"class"`
}


