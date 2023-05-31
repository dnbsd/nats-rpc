package natsrpc

type pipeline struct {
	subject   string
	group     string
	consumer  *consumer
	publisher *publisher
	service   Service
}
