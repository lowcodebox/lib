package logbox_client

import (
	"context"
	"fmt"
	"time"

	pb "git.lowcodeplatform.net/fabric/logbox/pkg/model/sdk"
)

func (c *client) Upsert(ctx context.Context, in upsertReq) (out upsertRes, err error) {
	conn, err := c.client.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("cannot get grpc connection. err: %s, client: %+v", err, c.client)
		return out, err
	}
	if conn == nil {
		err = fmt.Errorf("cannot get grpc connection (connection is null)")
		return out, err
	}

	// добавил выход по контексту, для случаев, если соединение таймаутит
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	events := &pb.UpsertRequest{}
	for _, v := range in.Events {
		events.Events = append(events.Events, &pb.Event{
			Config: v.config,
			Level:  v.level,
			Uid:    v.uid,
			Time:   v.time,
			Srv:    v.srv,
			Msg:    v.msg,
			Name:   v.name,
		})
	}

	client := pb.NewLogboxClient(conn)
	res, err := client.Upsert(ctxWithDeadline, events)
	out.Status = res.Status
	out.Error = res.Error

	if err != nil {
		return out, err
	}
	if res.Error != "" {
		err = fmt.Errorf("error send message. err: %s", res.Error)
	}
	if !res.Status {
		err = fmt.Errorf("fail status. err: %s", err)
	}

	return out, err
}
