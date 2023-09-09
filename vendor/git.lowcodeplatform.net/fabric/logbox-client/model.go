package logbox_client

type event struct {
	Uid       string `json:"uid"`
	Level     string `json:"level"`
	Name      string `json:"logger"`
	Type      string `json:"service-type"`
	Time      string `json:"ts"`
	Timing    string `json:"timing"`
	ConfigID  string `json:"config-id"`
	RequestID string `json:"request-id"`
	ServiceID string `json:"service-id"`
	Msg       string `json:"msg"`
	Payload   string `json:"payload"`
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
	Uid string,
	Level string,
	Type string,
	Name string,
	ConfigID string,
	RequestID string,
	ServiceID string,
	Msg string,
	Time string,
	Timing string,
	Payload string,
) *event {
	return &event{
		Level:     Level,
		Msg:       Msg,
		Name:      Name,
		ConfigID:  ConfigID,
		ServiceID: ServiceID,
		RequestID: RequestID,
		Time:      Time,
		Uid:       Uid,
		Type:      Type,
		Timing:    Timing,
		Payload:   Payload,
	}
}
