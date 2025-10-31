package queue

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
)

// RabbitMQ RabbitMQ 客户端
type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	config    config.RabbitMQConfig
	reconnect chan bool
}

// MQClient 全局 RabbitMQ 客户端实例
var MQClient *RabbitMQ

// Init 初始化 RabbitMQ 连接
// 参数:
//
//	cfg: RabbitMQ 配置
//
// 返回:
//
//	error: 错误信息
func Init(cfg config.RabbitMQConfig) error {
	mq := &RabbitMQ{
		config:    cfg,
		reconnect: make(chan bool),
	}

	// 建立连接
	if err := mq.connect(); err != nil {
		return err
	}

	// 声明交换机和队列
	if err := mq.setup(); err != nil {
		return err
	}

	MQClient = mq

	// 启动重连监听
	go mq.handleReconnect()

	logger.Info("RabbitMQ 连接成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)

	return nil
}

// connect 建立连接
func (mq *RabbitMQ) connect() error {
	var err error

	// 连接 RabbitMQ
	mq.conn, err = amqp.Dial(mq.config.GetRabbitMQURL())
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	// 创建通道
	mq.channel, err = mq.conn.Channel()
	if err != nil {
		return fmt.Errorf("创建 RabbitMQ 通道失败: %w", err)
	}

	return nil
}

// setup 声明交换机和队列
func (mq *RabbitMQ) setup() error {
	// 声明交换机
	err := mq.channel.ExchangeDeclare(
		mq.config.Exchange.Name,
		mq.config.Exchange.Type,
		mq.config.Exchange.Durable,
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("声明交换机失败: %w", err)
	}

	// 声明队列并绑定
	for _, queueCfg := range mq.config.Queues {
		_, err := mq.channel.QueueDeclare(
			queueCfg.Name,
			queueCfg.Durable,
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("声明队列 %s 失败: %w", queueCfg.Name, err)
		}

		// 绑定队列到交换机
		err = mq.channel.QueueBind(
			queueCfg.Name,
			queueCfg.RoutingKey,
			mq.config.Exchange.Name,
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("绑定队列 %s 失败: %w", queueCfg.Name, err)
		}
	}

	return nil
}

// handleReconnect 处理自动重连
func (mq *RabbitMQ) handleReconnect() {
	for {
		reason, ok := <-mq.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			logger.Info("RabbitMQ 连接正常关闭")
			break
		}

		logger.Warn("RabbitMQ 连接断开，准备重连",
			zap.Error(reason),
		)

		// 尝试重连
		for {
			time.Sleep(5 * time.Second)
			if err := mq.connect(); err != nil {
				logger.Error("RabbitMQ 重连失败", zap.Error(err))
				continue
			}

			if err := mq.setup(); err != nil {
				logger.Error("RabbitMQ 设置失败", zap.Error(err))
				continue
			}

			logger.Info("RabbitMQ 重连成功")
			break
		}
	}
}

// Publish 发布消息
// 参数:
//
//	routingKey: 路由键
//	body: 消息内容
//
// 返回:
//
//	error: 错误信息
func (mq *RabbitMQ) Publish(routingKey string, body []byte) error {
	return mq.channel.Publish(
		mq.config.Exchange.Name,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

// Consume 消费消息
// 参数:
//
//	queueName: 队列名称
//	handler: 消息处理函数
//
// 返回:
//
//	error: 错误信息
func (mq *RabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := mq.channel.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack (手动确认)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("开始消费队列 %s 失败: %w", queueName, err)
	}

	// 处理消息
	go func() {
		for msg := range msgs {
			logger.Debug("收到消息",
				zap.String("queue", queueName),
				zap.String("routing_key", msg.RoutingKey),
			)

			// 处理消息
			if err := handler(msg.Body); err != nil {
				logger.Error("处理消息失败",
					zap.String("queue", queueName),
					zap.Error(err),
				)
				// 消息处理失败，拒绝并重新入队
				msg.Nack(false, true)
			} else {
				// 消息处理成功，确认
				msg.Ack(false)
			}
		}
	}()

	logger.Info("开始消费队列", zap.String("queue", queueName))
	return nil
}

// Close 关闭连接
func (mq *RabbitMQ) Close() error {
	if mq.channel != nil {
		if err := mq.channel.Close(); err != nil {
			return err
		}
	}
	if mq.conn != nil {
		return mq.conn.Close()
	}
	return nil
}

// Close 关闭 RabbitMQ 连接
func Close() error {
	if MQClient != nil {
		return MQClient.Close()
	}
	return nil
}
