package logbox_client

import "strings"

type event struct {
	Storage string
	Fields  []Field
	Payload string `json:"payload"`
}

type Field struct {
	Name  string
	Value string
}

type Column struct {
	Name        string
	SearchValue string
}

type setReq struct {
	Events []event
}

type setRes struct {
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

func (u *setReq) AddEvent(event event) {
	u.Events = append(u.Events, event)
}

func (c *client) NewEvent(storage string, fields ...Field) event {
	e := event{
		Storage: storage,
		Fields:  fields,
	}
	if len(fields) == 0 {
		e.Payload = "{}"
		return e
	}
	builder := strings.Builder{}
	builder.WriteRune('{')
	// Первое поле
	builder.WriteRune('"')
	builder.WriteString(fields[0].Name)
	builder.WriteString(`":"`)
	builder.WriteString(fields[0].Value)
	builder.WriteRune('"')
	// Следующие поля
	for _, f := range fields[1:] {
		builder.WriteString(`,"`)
		builder.WriteString(f.Name)
		builder.WriteString(`":"`)
		builder.WriteString(f.Value)
		builder.WriteRune('"')
	}
	builder.WriteRune('}')
	e.Payload = builder.String()

	return e
}
