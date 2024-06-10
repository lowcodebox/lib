package logbox_client

import (
	pb "git.edtech.vm.prod-6.cloud.el/wb/analyticscollector/pkg/model/sdk"
	"strings"
	"time"
)

type event struct {
	Storage string
	Fields  []Field
	Payload string `json:"payload"`
}

type Event struct {
	Storage   string
	Fields    []Field
	Timestamp time.Time
	Payload   string `json:"payload"`
}

type Field struct {
	Name  string
	Value string
}

type setReq struct {
	Events []event
}

type SetRes struct {
	UIDs   []string
	Status bool
	Error  string
}

type searchReq struct {
	*pb.SearchRequest
}

type SearchResponse struct {
	Data         []Event
	ResultSize   uint
	ResultLimit  int
	ResultOffset int
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

func (c *client) NewSearchReq(storage string, limit, offset int, fields ...Field) (sr searchReq) {
	sr.SearchRequest = &pb.SearchRequest{}
	sr.SearchRequest.Storage = storage
	sr.Offset = int64(offset)
	sr.Count = int64(limit)
	sr.Items = make([]*pb.EventItem, len(fields))
	for i, f := range fields {
		sr.Items[i] = &pb.EventItem{
			Field: f.Name,
			Value: f.Value,
		}
	}
	return
}

func (c *client) SearchWithOrder(sr searchReq, orderField string, asc bool) searchReq {
	sr.SearchRequest.OrderField = orderField
	sr.Asc = asc
	return sr
}
