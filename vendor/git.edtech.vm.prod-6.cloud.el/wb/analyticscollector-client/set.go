package logbox_client

import (
	"context"
	"fmt"
	"time"

	pb "git.edtech.vm.prod-6.cloud.el/wb/analyticscollector/pkg/model/sdk"
)

func (c *client) set(ctx context.Context, in setReq) (out setRes, err error) {
	conn, err := c.client.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("cannot get grpc connection. err: %s, client: %+v", err, c.client)
		return out, err
	}
	if conn == nil {
		err = fmt.Errorf("cannot get grpc connection (connection is null)")
		return out, err
	}

	startTime := time.Now()

	// добавил выход по контексту, для случаев, если соединение таймаутит
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	events := &pb.SetRequest{}
	for _, v := range in.Events {
		ev := &pb.Event{
			Storage: v.Storage,
			Payload: v.Payload,
			Items:   make([]*pb.EventItem, len(v.Fields)),
		}
		for i, f := range v.Fields {
			ev.Items[i] = &pb.EventItem{
				Field: f.Name,
				Value: f.Value,
			}
		}
		events.Events = append(events.Events, ev)
	}

	client := pb.NewCollectorClient(conn)
	res, err := client.Set(ctxWithDeadline, events)
	if err != nil {
		return out, fmt.Errorf("collector request timing: %f(c), err: %s", time.Since(startTime).Seconds(), err)
	}
	if res == nil {
		return out, fmt.Errorf("error set message to collector. result from set is empty")
	}

	out.Status = res.Status
	out.Error = res.Error

	if res.Error != "" {
		err = fmt.Errorf("error send message. err: %s", res.Error)
	}
	if !res.Status {
		err = fmt.Errorf("fail status. err: %s", err)
	}

	return out, err
}
