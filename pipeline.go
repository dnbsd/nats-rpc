package natsrpc

type pipeline struct {
	subject   string
	consumer  *consumer
	publisher *publisher
	service   Service
}
