package collector_client

import (
	"context"
	"fmt"
	pb "git.edtech.vm.prod-6.cloud.el/extensions/collector/pkg/model/sdk"
)

func (c *client) exec(ctx context.Context, service, method, params string) (result, id string, err error) {
	// Получение соединения
	conn, err := c.client.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("cannot get grpc connection. err: %w", err)
		return
	}
	if conn == nil {
		err = fmt.Errorf("cannot get grpc connection (connection is null)")
		return
	}
	// Тело запроса
	execIn := &pb.ExecIn{
		Service: service,
		Method:  method,
		Params:  params,
	}
	client := pb.NewCollectorClient(conn)
	// Таймаут
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	// Выполнение
	res, err := client.Exec(ctx, execIn)
	if err != nil {
		return
	}
	// Вернуть результат
	result = res.Result
	id = res.Id
	return
}
