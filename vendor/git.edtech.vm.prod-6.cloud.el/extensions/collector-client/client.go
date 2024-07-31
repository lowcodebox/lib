package collector_client

import (
	"context"
	"fmt"
	"time"

	"git.edtech.vm.prod-6.cloud.el/packages/grpcbalancer"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var timeoutDefault = 2 * time.Second

type client struct {
	client     *grpcbalancer.Client
	timeout    time.Duration
	domain     string
	projectKey string
}

type Client interface {
	Exec(ctx context.Context, service, method, params string) (result, id string, err error)
	Close() error
}

func (c *client) Close() error {
	return c.client.Close()
}

func New(ctx context.Context, urlstr string, reqTimeout time.Duration, projectKey string) (Client, error) {
	if reqTimeout == 0 {
		reqTimeout = timeoutDefault
	}
	b, err := grpcbalancer.New(
		grpcbalancer.WithUrls(urlstr),
		grpcbalancer.WithInsecure(),
		grpcbalancer.WithTimeout(reqTimeout),
		grpcbalancer.WithChainUnaryInterceptors(GRPCUnaryClientInterceptor),
	)
	if err != nil {
		fmt.Printf("failed init grpcbalancer, err: %s", err)

		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("error init connect (grpcbalancer) client")
	}
	// фикс, если нет соединения, то возвращается {}
	if fmt.Sprint(b) == "{}" {
		return nil, fmt.Errorf("error init connect (grpcbalancer) client")
	}

	//u, err := url.Parse(urlstr)
	//if err != nil {
	//	return nil, err
	//}
	//splitUrl := strings.Split(u.Path, "/")
	//if len(splitUrl) < 3 {
	//	domain := []string{""}
	//}
	//splitUrl = splitUrl[1:3]

	return &client{
		client:     b,
		domain:     "", //strings.Join(splitUrl, "/"),
		projectKey: projectKey,
		timeout:    reqTimeout,
	}, nil
}

func GRPCUnaryClientInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var requestID string

	requestID = logger.GetRequestIDCtx(ctx)

	if requestID == "" {
		requestID = uuid.New().String()
	}

	ctx = logger.AddToGRPCHeader(ctx, "x-request-id", requestID)

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func (c *client) Exec(ctx context.Context, service, method, params string) (result, id string, err error) {
	return c.exec(ctx, service, method, params)
}
