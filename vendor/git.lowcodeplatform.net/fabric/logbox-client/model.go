package logbox_client

type event struct {
	config string `json:"config"`
	level  string `json:"level"`
	msg    string `json:"msg"`
	name   string `json:"name"`
	srv    string `json:"srv"`
	time   string `json:"time"`
	uid    string `json:"uid"`
}

type Column struct {
	Name        string
	SearchValue string
}

type upsertReq struct {
	Events []event
}

type upsertRes struct {
	Status bool
	Error  string
}

type searchReq struct {
	Draw            int     `json:"draw"`
	RecordsTotal    int     `json:"recordsTotal"`
	RecordsFiltered int     `json:"recordsFiltered"`
	Data            []event `json:"data"`
}

type searchRes struct {
	Columns []Column
	Search  string `json:"search"`
	Draw    int    `json:"draw"`
	Limit   int    `json:"limit"`
	Skip    int    `json:"skip"` //nolint
	OrderBy int    `json:"order_by"`
	Dir     string `json:"dir"`
}

func (u *upsertReq) AddEvent(event event) {
	u.Events = append(u.Events, event)
	return
}

func (c *client) NewEvent(
	Config string,
	Level string,
	Msg string,
	Name string,
	Srv string,
	Time string,
	Uid string,
) *event {
	return &event{
		config: Config,
		level:  Level,
		msg:    Msg,
		name:   Name,
		srv:    Srv,
		time:   Time,
		uid:    Uid,
	}
}
