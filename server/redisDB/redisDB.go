package redisDB

import (
	"context"
	"easy-chat/server/object"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisHandler struct {
	rdb *redis.Client
}

func NewRedisHandler(config object.Config) *RedisHandler {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + config.Redis.Port,
		Password: config.Redis.Pwd,
		DB:       config.Redis.Db,
	})
	return &RedisHandler{
		rdb: rdb,
	}
}

func (r *RedisHandler) MsgQueuePop(ctx context.Context) ([]string, error) {
	result, err := r.rdb.BLPop(ctx, 0*time.Second, "message_queue").Result()
	return result, err
}

func (r *RedisHandler) MsgQueuePush(ctx context.Context, msg string) error {
	err := r.rdb.RPush(ctx, "message_queue", msg).Err()
	return err
}

func (r *RedisHandler) AddScore(ctx context.Context, nickName string) error {
	err := r.rdb.ZIncrBy(ctx, "user_activity", 1, nickName).Err()
	return err
}

func (r *RedisHandler) ShowRank(ctx context.Context) (string, error) {
	zs, err := r.rdb.ZRangeWithScores(ctx, "user_activity", 0, -1).Result()
	if err != nil {
		return "", errors.New("获取用户活跃度失败: " + err.Error())
	}

	// 检查有序集合是否为空
	if len(zs) == 0 {
		return "", errors.New("排行榜为空")
	}

	// 返回排名、用户和分数
	msg := "用户活跃度排行榜:\n"
	for i, z := range zs {
		msg += fmt.Sprintf("%d. %s  积分: %.0f", i+1, z.Member, z.Score)
		if i < len(zs)-1 {
			msg += "\n"
		}
	}
	return msg, nil
}

func (r *RedisHandler) Clean(ctx context.Context) error {
	// 删除 message_queue 队列
	err := r.rdb.Del(ctx, "message_queue").Err()
	if err != nil {
		return errors.New("清理消息队列失败: " + err.Error())
	}
	// 删除 user_activity 有序集合
	err = r.rdb.Del(ctx, "user_activity").Err()
	if err != nil {
		return errors.New("清理用户活跃度失败: " + err.Error())
	}
	return nil
}
