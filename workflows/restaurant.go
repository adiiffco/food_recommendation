package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	r "foodapp/adapters/redis"
	c "foodapp/constants"
	m "foodapp/models"
	u "foodapp/utils"
)

func storeRestaurantList(ctx context.Context, restaurantList []m.Restaurant) error {
	for _, res := range restaurantList {
		err := storeRestaurantInfo(ctx, res)
		if err != nil {
			return err
		}
	}
	return nil
}

func storeRestaurantInfo(ctx context.Context, res m.Restaurant) error {
	//Add restaurant details
	bs, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = r.Rdb.Set(ctx, fmt.Sprintf(c.RESTAURANT_INFO, res.Id), string(bs), 0).Result()
	if err != nil {
		return err
	}
	//Add to master list
	err = r.AddToSet(ctx, c.ALL_RECOMMENDATION_KEY, res.Id)
	if err != nil {
		return err
	}
	//Add Recommeded restaurants
	if res.IsRecommended {
		err := r.AddToSet(ctx, c.RECOMMENDED_RESTAURANTS_KEY, res.Id)
		if err != nil {
			return err
		}
	}
	//Add restaurant cuisines
	err = r.AddToSet(ctx, fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, res.Cuisine), res.Id)
	if err != nil {
		return err
	}
	//Add restaurant cost types
	err = r.AddToSet(ctx, fmt.Sprintf(c.COST_RESTAURANT_KEY, res.CostBracket), res.Id)
	if err != nil {
		return err
	}
	//Add New onboarded restaurants
	if u.GetTimeDifferenceInHour(res.OnBoardedTime) <= c.NEW_RESTAURANT_HOUR_LIMIT {
		err = r.AddToZSet(ctx, c.NEW_RESTAURANTS_KEY, res.Id, res.Rating)
		if err != nil {
			return err
		}
	}
	//Add restaurant rating
	err = r.AddToSet(ctx, fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(res.Rating)), res.Id)
	if err != nil {
		return err
	}
	return nil
}

func getRestaurantDetails(ctx context.Context, list []string) ([]m.Restaurant, error) {
	result := make([]m.Restaurant, 0)
	for i := len(list) - 1; i >= 0; i-- {
		resID := list[i]
		temp := m.Restaurant{}
		res, err := r.Rdb.Get(ctx, fmt.Sprintf(c.RESTAURANT_INFO, resID)).Result()
		if err != nil {
			return result, err
		}
		bs := []byte(res)
		err = json.Unmarshal(bs, &temp)
		if err != nil {
			return result, err
		}
		result = append(result, temp)
	}
	return result, nil
}
