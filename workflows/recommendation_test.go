package workflows

import (
	"fmt"
	r "foodapp/adapters/redis"
	c "foodapp/constants"
	m "foodapp/models"
	"testing"
	"time"
)

func TestRecommendation(t *testing.T) {
	r.Initialize()
	userFilter := m.UserTracking{
		Cuisines: []m.CuisineTracking{
			{Type: c.SouthIndian, NoOfOrders: 15},
			{Type: c.NorthIndian, NoOfOrders: 5},
			{Type: c.Chinese, NoOfOrders: 10},
		},
		Costs: []m.CostTracking{
			{Type: 1, NoOfOrders: 5},
			{Type: 2, NoOfOrders: 7},
			{Type: 3, NoOfOrders: 11},
			{Type: 4, NoOfOrders: 1},
		},
	}

	restaurantList := []m.Restaurant{
		{Id: 2, Cuisine: c.NorthIndian, CostBracket: 4,
			Rating: 3.7, IsRecommended: false, OnBoardedTime: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)},
		{Id: 67, Cuisine: c.Chinese, CostBracket: 1,
			Rating: 4.8, IsRecommended: false, OnBoardedTime: time.Date(2022, 8, 15, 14, 30, 45, 100, time.Local)},
		{Id: 79, Cuisine: c.SouthIndian, CostBracket: 3,
			Rating: 4, IsRecommended: true, OnBoardedTime: time.Date(2023, 4, 12, 14, 30, 45, 100, time.Local)},
	}
	restList := GetRestaurantRecommendation(userFilter, restaurantList)
	fmt.Println("response: ", restList)
	if len(restList) > 0 {
		t.Log("Passed")
	}
}
