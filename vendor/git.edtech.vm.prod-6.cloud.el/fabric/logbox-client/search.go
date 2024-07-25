package logbox_client

import (
	"context"
	"fmt"
	"time"

	pb "git.edtech.vm.prod-6.cloud.el/fabric/logbox/pkg/model/sdk"
)

func (c *client) search(ctx context.Context, in searchRes) (out searchReq, err error) {
	//token, err := lib.GenXServiceKey(c.domain, []byte(c.projectKey), tokenInterval)
	//ctx = AddToGRPCHeader(ctx, headerServiceKey, token)

	conn, err := c.client.Conn(ctx)
	if err != nil || conn == nil {
		err = fmt.Errorf("[client] [logbox] cannot get grpc connection")
		return out, err
	}

	// добавили выход по контексту, для случаев, если соединение таймаутит
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	columns := []*pb.Column{}
	startTime := time.Now()

	for _, v := range in.Columns {
		column := pb.Column{
			Name:        v.Name,
			SearchValue: v.SearchValue,
		}
		columns = append(columns, &column)
	}

	inClient := &pb.SearchRequest{
		Columns: columns,
		Search:  in.Search,
		Draw:    uint64(in.Draw),
		Limit:   uint64(in.Limit),
		Skip:    uint64(in.Skip),
		OrderBy: uint64(in.OrderBy),
		Dir:     in.Dir,
	}

	client := pb.NewLogboxClient(conn)

	result, err := client.Search(ctxWithDeadline, inClient)
	if err != nil {
		return out, fmt.Errorf("[client] [logbox] request error request Search. timing: %d, err: %dms", time.Since(startTime).Milliseconds(), err)
	}
	if result == nil {
		return out, fmt.Errorf("[client] [logbox] error request Search. result is empty")
	}

	out.Draw = int(result.Draw)
	out.RecordsTotal = int(result.RecordsTotal)
	out.RecordsFiltered = int(result.RecordsFiltered)
	out.Data = []event{}
	for _, m := range result.Data {
		ev := c.NewEvent(
			m.Uid,
			m.Level,
			m.Type,
			m.Name,
			m.Config,
			m.Request,
			m.User,
			m.Srv,
			m.Msg,
			m.Time,
			m.Timing,
			m.Payload,
		)
		out.Data = append(out.Data, ev)
	}

	return out, err
}
