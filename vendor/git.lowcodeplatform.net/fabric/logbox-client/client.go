package logbox_client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.lowcodeplatform.net/packages/grpcbalancer"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/sony/gobreaker"
)

const (
	requestIDField   string = "request-id"
	headerServiceKey        = "X-Service-Key"
	tokenInterval           = 1 * time.Minute
)

var timeoutDefault = 1 * time.Second

type client struct {
	client     *grpcbalancer.Client
	cb         *gobreaker.CircuitBreaker
	domain     string
	projectKey string
}

type Client interface {
	Upsert(ctx context.Context, in upsertReq) (out upsertRes, err error)
	Search(ctx context.Context, in searchRes) (out searchReq, err error)

	NewUpsertReq() upsertReq
	NewEvent(Uid, Level, Type, Name, ConfigID, RequestID, UserID, ServiceID, Msg, Time, Timing, Payload string) event

	Close() error
}

func (c *client) NewUpsertReq() upsertReq {
	return upsertReq{}
}

func (c *client) Close() error {
	return c.client.Close()
}

func New(ctx context.Context, urlstr string, reqTimeout time.Duration, cbMaxRequests uint32, cbTimeout, cbInterval time.Duration, projectKey string) (Client, error) {
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

	if cbMaxRequests == 0 {
		cbMaxRequests = 3
	}
	if cbTimeout == 0 {
		cbTimeout = 5 * time.Second
	}
	if cbInterval == 0 {
		cbInterval = 5 * time.Second
	}

	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "apiCircuitBreaker",
			MaxRequests: cbMaxRequests, // максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии
			Timeout:     cbTimeout,     // период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние
			Interval:    cbInterval,    // циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				fmt.Errorf("apiCircuitBreaker is ReadyToTrip. counts.ConsecutiveFailures: %+v, err: %+v", zap.Any("counts.ConsecutiveFailures", counts.ConsecutiveFailures), zap.Error(err))
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				fmt.Errorf("apiCircuitBreaker changed position: name: %+v, from: %+v, to: %+v, err: %+v", zap.Any("name", name), zap.Any("from", from), zap.Any("to", to), zap.Error(err))
			},
		},
	)

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	splitUrl := strings.Split(u.Path, "/")
	if len(splitUrl) < 3 {
		return nil, fmt.Errorf("error path: %s", urlstr)
	}
	domain := splitUrl[1:3]

	return &client{
		client:     b,
		cb:         cb,
		domain:     strings.Join(domain, "/"),
		projectKey: projectKey,
	}, err
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

	requestID = GetRequestIDCtx(ctx, requestIDField)

	if requestID == "" {
		requestID = uuid.New().String()
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-request-id", requestID))

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func GetRequestIDCtx(ctx context.Context, name string) string {
	nameKey := "logger." + name
	requestID, _ := ctx.Value(nameKey).(string)

	return requestID
}
