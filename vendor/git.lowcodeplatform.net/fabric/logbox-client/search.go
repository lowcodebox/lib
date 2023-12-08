package logbox_client

import (
	"context"
	"fmt"
	"time"

	pb "git.lowcodeplatform.net/fabric/logbox/pkg/model/sdk"
)

func (c *client) Search(ctx context.Context, in searchRes) (out searchReq, err error) {
	conn, err := c.client.Conn(ctx)
	if err != nil || conn == nil {
		err = fmt.Errorf("[client] [logbox] cannot get grpc connection")
		return out, err
	}

	// добавили выход по контексту, для случаев, если соединение таймаутит
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	columns := []*pb.Column{}

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
		err = fmt.Errorf("[client] [logbox] error request Search. err: %s", err)
		return out, err
	}

	if result == nil {
		return out, fmt.Errorf("[client] [logbox] error request Search. result is empty")
	}

	out.Draw = int(result.Draw)
	out.RecordsTotal = int(result.RecordsTotal)
	out.RecordsFiltered = int(result.RecordsFiltered)
	out.Data = []event{}
	for _, m := range result.Data {
		ev := c.NewEvent(m.Config, m.Level, m.Msg, m.Name, m.Srv, m.Time, m.Uid)
		out.Data = append(out.Data, *ev)
	}

	return out, err
}