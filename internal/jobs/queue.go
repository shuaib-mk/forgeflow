package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgeflow/forgeflow/pkg/models"
	"github.com/redis/go-redis/v9"
)

const queueKey="forgeflow:workflow-runs"

type Queue interface{Enqueue(context.Context,models.ID)error;Dequeue(context.Context,time.Duration)(models.ID,error)}
type RedisQueue struct{client *redis.Client}

func NewRedisQueue(redisURL string)(*RedisQueue,error){options,err:=redis.ParseURL(redisURL);if err!=nil{return nil,fmt.Errorf("parse redis URL: %w",err)};return &RedisQueue{client:redis.NewClient(options)},nil}
func (q *RedisQueue) Close()error{return q.client.Close()}
func (q *RedisQueue) Ping(ctx context.Context)error{return q.client.Ping(ctx).Err()}
func (q *RedisQueue) Enqueue(ctx context.Context,id models.ID)error{if err:=q.client.LPush(ctx,queueKey,string(id)).Err();err!=nil{return fmt.Errorf("enqueue workflow run: %w",err)};return nil}
func (q *RedisQueue) Dequeue(ctx context.Context,timeout time.Duration)(models.ID,error){result,err:=q.client.BRPop(ctx,timeout,queueKey).Result();if errors.Is(err,redis.Nil){return "",context.DeadlineExceeded};if err!=nil{return "",fmt.Errorf("dequeue workflow run: %w",err)};return models.ID(result[1]),nil}

