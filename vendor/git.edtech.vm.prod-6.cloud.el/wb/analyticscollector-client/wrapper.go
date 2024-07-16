package logbox_client

import (
	"context"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

func (c *client) Set(ctx context.Context, in setReq) (out SetRes, err error) {
	return c.set(ctx, in)
}

func (c *client) Search(ctx context.Context, in searchReq) (out models.ResponseData, err error) {
	sOut, err := c.search(ctx, in)
	if err != nil {
		return
	}

	out.Data = make([]models.Data, len(sOut.Data))

	for i, data := range sOut.Data {
		d := models.Data{Attributes: make(map[string]models.Attribute)}

		for _, v := range data.Fields {
			d.Attributes[v.Name] = models.Attribute{Value: v.Value}
		}

		d.Attributes["timestamp"] = models.Attribute{Value: data.Timestamp.String()}
		d.Attributes["payload"] = models.Attribute{Value: data.Payload}

		out.Data[i] = d
	}

	out.Metrics = models.Metrics{
		ResultCount:  int(sOut.ResultSize),
		ResultOffset: sOut.ResultOffset,
		ResultLimit:  sOut.ResultLimit,
	}

	return
}

func (c *client) Query(ctx context.Context, uid string, offset int, params ...interface{}) (out models.ResponseData, err error) {
	qOut, err := c.query(ctx, uid, offset, params...)
	// преобразование к ResponseData
	if err != nil {
		return
	}

	out.Data = make([]models.Data, len(qOut))

	for i := range qOut {
		d := models.Data{
			Attributes: make(map[string]models.Attribute),
		}

		for k, v := range qOut[i] {
			d.Attributes[k] = models.Attribute{Value: fmt.Sprint(v)}
		}

		out.Data[i] = d
	}

	out.Metrics = models.Metrics{
		ResultCount:  len(qOut),
		ResultOffset: offset,
	}

	return
}
