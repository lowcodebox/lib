package logbox_client

import (
	"context"
	"fmt"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/grpcbalancer"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	defaultConnectTimeout = 1 * time.Second
	defaultRequestTimeout = 5 * time.Second
	defaultRetryInterval  = 5 * time.Second
	defaultRetries        = 2000
)

type client struct {
	client        *grpcbalancer.Client
	timeout       time.Duration
	domain        string
	projectKey    string
	setRetries    int
	retryInterval time.Duration
}

type Client interface {
	Set(ctx context.Context, in setReq) (out SetRes, err error)
	SetAsync(ctx context.Context, in setReq)
	Search(ctx context.Context, in searchReq) (out models.ResponseData, err error)
	Query(ctx context.Context, uid string, offset int, params ...interface{}) (out models.ResponseData, err error)

	NewSetReq() setReq
	NewEvent(storage string, fields ...Field) event

	NewSearchReq(storage string, limit, offset int, fields ...Field) searchReq
	SearchWithOrder(sr searchReq, orderField string, asc bool) searchReq

	Close() error
}

func (c *client) NewSetReq() setReq {
	return setReq{}
}

func (c *client) Close() error {
	return c.client.Close()
}

func New(ctx context.Context, urlstr string, reqTimeout time.Duration, projectKey string, opts ...Option) (Client, error) {
	if reqTimeout <= 0 {
		reqTimeout = defaultRequestTimeout
	}
	b, err := grpcbalancer.New(
		grpcbalancer.WithUrls(urlstr),
		grpcbalancer.WithInsecure(),
		grpcbalancer.WithTimeout(defaultConnectTimeout),
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

	c := &client{
		client:     b,
		domain:     "", //strings.Join(splitUrl, "/"),
		projectKey: projectKey,
		timeout:    reqTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.setRetries < 1 {
		c.setRetries = defaultRetries
	}

	if c.retryInterval <= 0 {
		c.retryInterval = defaultRetryInterval
	}
	if c.retryInterval <= c.timeout {
		c.retryInterval = c.timeout + time.Second
	}

	return c, nil
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
