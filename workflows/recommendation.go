package workflows

import (
	"context"
	"fmt"
	r "foodapp/adapters/redis"
	c "foodapp/constants"
	m "foodapp/models"
	u "foodapp/utils"
)

func GetRestaurantRecommendation(user m.UserTracking, restaurantList []m.Restaurant) []m.Restaurant {
	ctx := context.Background()
	res := make([]m.Restaurant, 0)
	err := storeUserOrderHistory(ctx, user)
	if err != nil {
		fmt.Println("error while storing order history for user: ", err)
		return res
	}
	cuisineTypes, err := getTopNCuisinesOfUser(ctx, c.CUISINE_TYPE_COUNT)
	if err != nil {
		fmt.Println("error while fetching cuisine types for user: ", err)
		return res
	}
	costTypes, err := getTopNCostTypesOfUser(ctx, c.COST_TYPE_COUNT)
	if err != nil {
		fmt.Println("error while fetching cost types for user: ", err)
		return res
	}
	userFilters := m.UserTopFilter{
		CuisineTypes: cuisineTypes,
		CostTypes:    costTypes,
	}
	err = storeRestaurantList(ctx, restaurantList)
	if err != nil {
		fmt.Println("error while storing restaurant list: ", err)
		return res
	}
	list, err := generateRecommendations(ctx, userFilters)
	if err != nil {
		fmt.Println("error in generating recommendation: ", err)
	}
	res, err = getRestaurantDetails(ctx, list)
	if err != nil {
		fmt.Println("error in getting recommendation details: ", err)
	}
	return res
}

func generateRecommendations(ctx context.Context, userFilters m.UserTopFilter) ([]string, error) {
	resultList := make([]string, 0)
	resultKey := c.RESULT_RECOMMENDATION_KEY
	featuredRestKey := c.RECOMMENDED_RESTAURANTS_KEY
	if len(userFilters.CuisineTypes) > 0 && len(userFilters.CostTypes) > 0 {
		//Featured + Cuisine1 + Cost1
		err := r.SInterWithLPush(ctx, resultKey, featuredRestKey,
			fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
			fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
		)
		if err != nil {
			return resultList, err
		}
		mem, err := r.LCard(ctx, resultKey)
		if err != nil {
			return resultList, err
		}
		if mem == 0 {
			if len(userFilters.CostTypes) > 1 {
				err := r.SInterWithLPush(ctx, resultKey, featuredRestKey,
					fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
					fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[1]),
				)
				if err != nil {
					return resultList, err
				}
			}
			if len(userFilters.CuisineTypes) > 1 {
				err := r.SInterWithLPush(ctx, resultKey, featuredRestKey,
					fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[1]),
					fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
				)
				if err != nil {
					return resultList, err
				}
			}
		}
		if checkIfRecommendationCompleted(ctx, &resultList) {
			return resultList, nil
		}
		//Cuisine1+cost1+rating>=4
		err = r.SInterWithLPush(ctx, resultKey,
			fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
			fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
			fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(4)),
		)
		if err != nil {
			return resultList, err
		}
		if checkIfRecommendationCompleted(ctx, &resultList) {
			return resultList, nil
		}
		//Cuisine1+cost2+rating>=4.5
		if len(userFilters.CostTypes) > 1 {
			err = r.SInterWithLPush(ctx, resultKey,
				fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
				fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[1]),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(4.5)),
			)
			if err != nil {
				return resultList, err
			}
			if checkIfRecommendationCompleted(ctx, &resultList) {
				return resultList, nil
			}
		}
		//Cuisine2+cost1+rating>=4.5
		if len(userFilters.CuisineTypes) > 1 {
			err = r.SInterWithLPush(ctx, resultKey,
				fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[1]),
				fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(4.5)),
			)
			if err != nil {
				return resultList, err
			}
			if checkIfRecommendationCompleted(ctx, &resultList) {
				return resultList, nil
			}
		}
		//Top 4 new restaurants based rating
		res, err := r.ZRevRangeByScore(ctx, c.NEW_RESTAURANTS_KEY, 0, 4)
		if err != nil {
			return resultList, err
		}
		r.AddToList(ctx, resultKey, res)
		if checkIfRecommendationCompleted(ctx, &resultList) {
			return resultList, nil
		}
		//Cuisine1+cost1+rating<4
		err = r.SInterWithLPush(ctx, resultKey,
			fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
			fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
			fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(3)),
		)
		if err != nil {
			return resultList, err
		}
		if checkIfRecommendationCompleted(ctx, &resultList) {
			return resultList, nil
		}
		//Cuisine1+cost2+rating<4.5
		if len(userFilters.CostTypes) > 1 {
			err = r.SInterWithLPush(ctx, resultKey,
				fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[0]),
				fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[1]),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(4)),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(3)),
			)
			if err != nil {
				return resultList, err
			}
			if checkIfRecommendationCompleted(ctx, &resultList) {
				return resultList, nil
			}
		}
		//Cuisine2+cost1+rating<4.5
		if len(userFilters.CuisineTypes) > 1 {
			err = r.SInterWithLPush(ctx, resultKey,
				fmt.Sprintf(c.CUISINE_RESTAURANT_KEY, userFilters.CuisineTypes[1]),
				fmt.Sprintf(c.COST_RESTAURANT_KEY, userFilters.CostTypes[0]),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(4)),
				fmt.Sprintf(c.RATING_RESTAURANT_KEY, u.GetKeyForRating(3)),
			)
			if err != nil {
				return resultList, err
			}
			if checkIfRecommendationCompleted(ctx, &resultList) {
				return resultList, nil
			}
		}
	}

	masterList, err := r.SMembers(ctx, c.ALL_RECOMMENDATION_KEY)
	if err != nil {
		return resultList, err
	}
	err = r.AddToListWithLimit(ctx, resultKey, masterList, c.TOTAL_RECOMMENDATION)
	if err != nil {
		return resultList, err
	}
	populateRecommendedRestaurants(ctx, &resultList)
	return resultList, nil
}

func checkIfRecommendationCompleted(ctx context.Context, resultList interface{}) bool {
	populateRecommendedRestaurants(ctx, resultList)
	count, err := r.LCard(ctx, c.RESULT_RECOMMENDATION_KEY)
	if err != nil {
		return false
	}
	if count >= c.TOTAL_RECOMMENDATION {
		return true
	}
	return false
}

func populateRecommendedRestaurants(ctx context.Context, result interface{}) {
	list, err := r.ListMembers(ctx, c.RESULT_RECOMMENDATION_KEY)
	if err != nil {
		return
	}
	u.ConvertToObj(list, result)
}
