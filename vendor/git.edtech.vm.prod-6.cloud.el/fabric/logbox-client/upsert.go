package logbox_client

import (
	"context"
	"fmt"
	"time"

	pb "git.edtech.vm.prod-6.cloud.el/fabric/logbox/pkg/model/sdk"
)

func (c *client) upsert(ctx context.Context, in upsertReq) (out upsertRes, err error) {
	//token, err := lib.GenXServiceKey(c.domain, []byte(c.projectKey), tokenInterval)
	//ctx = AddToGRPCHeader(ctx, headerServiceKey, token)

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

	events := &pb.UpsertRequest{}
	for _, v := range in.Events {
		events.Events = append(events.Events, &pb.Event{
			Config:  v.ConfigID,
			Level:   v.Level,
			Uid:     v.Uid,
			Time:    v.Time,
			Srv:     v.ServiceID,
			Msg:     v.Msg,
			Name:    v.Name,
			Timing:  v.Timing,
			Type:    v.Type,
			Request: v.RequestID,
			User:    v.UserID,
			Payload: v.Payload,
		})
	}

	client := pb.NewLogboxClient(conn)
	res, err := client.Upsert(ctxWithDeadline, events)
	if err != nil {
		return out, fmt.Errorf("request timing: %f(c), err: %s", time.Since(startTime).Seconds(), err)
	}
	if res == nil {
		return out, fmt.Errorf("error upsert message to logbox. result from upsert is empty")
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
