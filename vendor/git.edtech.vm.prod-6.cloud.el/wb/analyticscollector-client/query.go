package logbox_client

import (
	"context"
	"fmt"
	pb "git.edtech.vm.prod-6.cloud.el/wb/analyticscollector/pkg/model/sdk"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *client) query(ctx context.Context, uid string, offset int, params ...interface{}) (out QueryResult, err error) {
	conn, err := c.client.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("cannot get grpc connection. err: %w, client: %+v", err, c.client)
		return out, err
	}
	if conn == nil {
		err = fmt.Errorf("cannot get grpc connection (connection is null)")
		return out, err
	}
	paramsList, err := structpb.NewList(params)
	if err != nil {
		return out, err
	}

	client := pb.NewCollectorClient(conn)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := client.Query(ctx, &pb.QueryRequest{
		UID:    uid,
		Offset: int64(offset),
		Params: paramsList,
	})

	if err != nil {
		return out, err
	}
	out = make(QueryResult, len(res.Data))
	for i := range res.Data {
		out[i] = res.Data[i].AsMap()
	}

	return out, nil
}
