package logbox_client

import (
	"context"
	"fmt"
	pb "git.edtech.vm.prod-6.cloud.el/wb/analyticscollector/pkg/model/sdk"
	"time"
)

func (c *client) search(ctx context.Context, in searchReq) (out SearchResponse, err error) {
	conn, err := c.client.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("cannot get grpc connection. err: %w, client: %+v", err, c.client)
		return out, err
	}
	if conn == nil {
		err = fmt.Errorf("cannot get grpc connection (connection is null)")
		return out, err
	}

	ctxWithDeadline, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	collectorClient := pb.NewCollectorClient(conn)
	res, err := collectorClient.Search(ctxWithDeadline, in.SearchRequest)
	if err != nil {
		return out, fmt.Errorf("error search message to collector. err: %w", err)
	}
	if res == nil {
		return out, fmt.Errorf("error search message to collector. (result is null)")
	}
	if res.Error != "" {
		return out, fmt.Errorf("error search message. err: %s", res.Error)
	}

	out.ResultSize = uint(res.Metrics.ResultCount)
	out.ResultOffset = int(res.Metrics.ResultOffset)
	out.ResultLimit = int(res.Metrics.ResultLimit)

	out.Data = make([]Event, len(res.Events))
	// Заполнение эвентов
	for i, e := range res.Events {
		out.Data[i] = Event{
			Storage:   e.Storage,
			Payload:   e.Payload,
			Timestamp: e.Timestamp.AsTime(),
			Fields:    make([]Field, len(e.Items)),
		}
		// Заполнение полей эвента
		for j, item := range e.Items {
			out.Data[i].Fields[j] = Field{
				Name:  item.Field,
				Value: item.Value,
			}
		}
	}

	return
}
