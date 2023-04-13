package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func Initialize() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}
	rdb.FlushAll(ctx).Result()
	Rdb = rdb
}

func AddToSet(ctx context.Context, key string, members ...interface{}) error {
	_, err := Rdb.SAdd(ctx, key, members...).Result()
	return err
}

func ExistsInList(ctx context.Context, key, value string) bool {
	pos, err := Rdb.LPos(ctx, key, value, redis.LPosArgs{}).Result()
	if err != nil || pos == -1 {
		return false
	}
	return true
}

func AddToList(ctx context.Context, key string, members []string) error {
	for _, m := range members {
		if !ExistsInList(ctx, key, m) {
			_, err := Rdb.LPush(ctx, key, m).Result()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AddToListWithLimit(ctx context.Context, key string, members []string, limit int64) error {
	for _, m := range members {
		card, _ := LCard(ctx, key)
		if card == limit {
			break
		}
		if !ExistsInList(ctx, key, m) {
			_, err := Rdb.LPush(ctx, key, m).Result()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AddToZSet(ctx context.Context, key string, member interface{}, score float64) error {
	_, err := Rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	return err
}

func SInterStore(ctx context.Context, destKey string, keys ...string) error {
	_, err := Rdb.SInterStore(ctx, destKey, keys...).Result()
	return err
}

func SInterWithLPush(ctx context.Context, destKey string, keys ...string) error {
	result, err := Rdb.SInter(ctx, keys...).Result()
	if err != nil {
		return err
	}
	if len(result) > 0 {
		err = AddToList(ctx, destKey, result)
	}
	return err
}

func SDiffStore(ctx context.Context, destKey string, keys ...string) error {
	_, err := Rdb.SDiffStore(ctx, destKey, keys...).Result()
	return err
}

func SDiffWithLPush(ctx context.Context, destKey string, keys ...string) error {
	result, err := Rdb.SDiff(ctx, keys...).Result()
	if err != nil {
		return err
	}
	if len(result) > 0 {
		err = AddToList(ctx, destKey, result)
	}
	return err
}

func SMemCount(ctx context.Context, key string) (int64, error) {
	return Rdb.SCard(ctx, key).Result()
}

func SMemExists(ctx context.Context, key string) bool {
	mcount, err := SMemCount(ctx, key)
	if err != nil || mcount == 0 {
		return false
	}
	return true
}

func ZRevRangeByScore(ctx context.Context, key string, offset, count int64) ([]string, error) {
	res, err := Rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Offset: offset,
		Count:  count,
		Min:    "0",
		Max:    "+inf",
	}).Result()
	return res, err
}

func SMembers(ctx context.Context, key string) ([]string, error) {
	return Rdb.SMembers(ctx, key).Result()
}

func LCard(ctx context.Context, key string) (int64, error) {
	return Rdb.LLen(ctx, key).Result()
}

func ListMembers(ctx context.Context, key string) ([]string, error) {
	len, err := LCard(ctx, key)
	if err != nil {
		return []string{}, err
	}
	return Rdb.LRange(ctx, key, 0, len).Result()
}
