package msg_queue

import (
	"context"
	"fmt"
	"time"

	"platform/logger"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-json"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

type (
	HandleFunc = func(*kgo.Record)
	Handler    struct {
		fn HandleFunc

		isAckBeforeProcessing bool
	}
	HandlerTable = map[string]Handler
)

type kafkaConfig struct {
	Addrs    string `env:"KAFKA_ADDRS, required, notEmpty"`
	Topics   string `env:"KAFKA_TOPICS, required, notEmpty"`
	Group    string `env:"KAFKA_GROUP, required, notEmpty"`
	User     string `env:"KAFKA_USER, required, notEmpty"`
	Password string `env:"KAFKA_PASSWORD, required, notEmpty"`
}

type KGOClient struct {
	client *kgo.Client
	table  HandlerTable
}

func MustCreateKafkaClient(
	table HandlerTable,
	maxGoroutines uint,
) *KGOClient {
	cfg, err := env.ParseAs[kafkaConfig]()
	if err != nil {
		logger.Fatal(err.Error())

		return nil
	}

	var addrsArr []string

	if err = json.Unmarshal([]byte(cfg.Addrs), &addrsArr); err != nil {
		logger.Fatal(fmt.Sprintf("error occurred when unmarshalling kafka Addrs: %v", err))

		return nil
	}

	if len(addrsArr) == 0 {
		logger.Fatal("error occurred when creating a kafka client: empty address")

		return nil
	}

	var topics []string

	if err = json.Unmarshal([]byte(cfg.Topics), &addrsArr); err != nil {
		logger.Fatal(fmt.Sprintf("error occurred when unmarshalling kafka Topics: %v", err))

		return nil
	}

	for _, topic := range topics {
		if _, ok := table[topic]; !ok {
			logger.Fatal(
				fmt.Sprintf(
					"error occurred when creating a kafka client: topic %s is not in the handle table",
					topic,
				),
			)
		}
	}

	kgoClient, err := kgo.NewClient(
		kgo.SeedBrokers(addrsArr...),
		kgo.ConsumerGroup(cfg.Group),
		kgo.ConsumeTopics(cfg.Topics),
		kgo.FetchMinBytes(1<<10),
		kgo.FetchMaxBytes(4<<20),
		kgo.FetchMaxWait(time.Millisecond),
		kgo.DisableAutoCommit(),
		kgo.SASL(plain.Auth{
			User: cfg.User,
			Pass: cfg.Password,
		}.AsMechanism()),
		kgo.SessionTimeout(30*time.Second),
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error occurred when creating a kafka client: %v", err))
	}

	client := &KGOClient{
		client: kgoClient,
		table:  table,
	}

	if len(topics) > 0 {
		go client.startPolling(maxGoroutines)
	}

	return client
}

func (client *KGOClient) ack(record *kgo.Record) {
	if err := client.client.CommitRecords(client.client.Context(), record); err == nil {
		logger.Errorf("error occurred when committing record: %v", err)
	}
}

// startPolling does not run a new goroutine and should be called in a new one.
func (client *KGOClient) startPolling(maxGoroutines uint) {
	if maxGoroutines == 0 {
		maxGoroutines = 10000
	}

	type workerArgs struct {
		client *KGOClient
		record *kgo.Record
	}

	pool, err := ants.NewPoolWithFuncGeneric(
		int(maxGoroutines),
		func(args workerArgs) {
			record := args.record
			table := args.client.table
			handler := table[record.Topic]

			if handler.isAckBeforeProcessing {
				args.client.ack(record)
			}

			handler.fn(args.record)

			if !handler.isAckBeforeProcessing {
				args.client.ack(record)
			}
		},
		ants.WithLogger(logger.MainLogger()),
		ants.WithPanicHandler(func(err interface{}) {
			logger.Fatalf("panic occurred in kafka worker: %v", err)
		}),
		ants.WithExpiryDuration(time.Minute),
		ants.WithNonblocking(false),
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error occurred when creating a kafka worker pool: %v", err))
	}

	ctx := client.client.Context()

	for ctx.Err() == nil {
		fetches := client.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				logger.Errorf("error occurred when polling fetches: %v", e)
			}

			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			err = pool.Invoke(workerArgs{
				client: client,
				record: record,
			})
			if err != nil {
				logger.Errorf("error occurred when invoking a kafka worker: %v", err)
			}
		})
	}

	pool.Release()
}

func (client *KGOClient) Produce(ctx context.Context, topic string, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Value: value,
	}

	if err := client.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return errors.Wrap(err, "error occurred when producing a record")
	}

	return nil
}

func (client *KGOClient) Close() {
	client.client.Close()
}
