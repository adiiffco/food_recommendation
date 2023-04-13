package workflows

import (
	"context"
	r "foodapp/adapters/redis"
	c "foodapp/constants"
	m "foodapp/models"
	u "foodapp/utils"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func storeUserOrderHistory(ctx context.Context, userOrderHistory m.UserTracking) error {
	cuisineList := make([]redis.Z, len(userOrderHistory.Cuisines))
	for i, cuisine := range userOrderHistory.Cuisines {
		var choice int
		err := u.ConvertToObj(cuisine.Type, &choice)
		if err != nil {
			return err
		}
		cuisineList[i].Member = choice
		cuisineList[i].Score = float64(cuisine.NoOfOrders)
	}
	_, err := r.Rdb.ZAdd(ctx, c.USER_CUISINES_LIST, cuisineList...).Result()
	if err != nil {
		return err
	}

	costTypeList := make([]redis.Z, len(userOrderHistory.Costs))
	for i, cost := range userOrderHistory.Costs {
		costTypeList[i].Member = cost.Type
		costTypeList[i].Score = float64(cost.NoOfOrders)
	}
	_, err = r.Rdb.ZAdd(ctx, c.USER_COST_TYPE_LIST, costTypeList...).Result()
	if err != nil {
		return err
	}
	return nil
}

func getTopNCuisinesOfUser(ctx context.Context, n int64) ([]c.CuisineChoices, error) {
	cuisineTypes := make([]c.CuisineChoices, 0)
	res, err := r.ZRevRangeByScore(ctx, c.USER_CUISINES_LIST, 0, n)
	if err != nil {
		return cuisineTypes, err
	}
	for _, cType := range res {
		num, err := strconv.Atoi(cType)
		if err != nil {
			return cuisineTypes, err
		}
		var choice c.CuisineChoices
		err = u.ConvertToObj(num, &choice)
		if err != nil {
			return cuisineTypes, err
		}
		cuisineTypes = append(cuisineTypes, choice)
	}
	return cuisineTypes, nil
}

func getTopNCostTypesOfUser(ctx context.Context, n int64) ([]int, error) {
	costTypes := make([]int, 0)
	res, err := r.ZRevRangeByScore(ctx, c.USER_COST_TYPE_LIST, 0, n)
	if err != nil {
		return costTypes, err
	}
	for _, cType := range res {
		costType, err := strconv.Atoi(cType)
		if err != nil {
			return costTypes, err
		}
		costTypes = append(costTypes, costType)
	}
	return costTypes, nil
}
