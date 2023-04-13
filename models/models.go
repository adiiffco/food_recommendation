package models

import (
	c "foodapp/constants"
	"time"
)

type Restaurant struct {
	Id            int              `json:"restaurantId"`
	Cuisine       c.CuisineChoices `json:"cuisine"`
	CostBracket   int              `json:"costBracket"`
	Rating        float64          `json:"rating"`
	IsRecommended bool             `json:"isRecommended"`
	OnBoardedTime time.Time        `json:"onboardedTime"`
}

type CuisineTracking struct {
	Type       c.CuisineChoices `json:"type"`
	NoOfOrders int              `json:"noOfOrders"`
}

type CostTracking struct {
	Type       int `json:"type"`
	NoOfOrders int `json:"noOfOrders"`
}

type UserTracking struct {
	Cuisines []CuisineTracking `json:"cuisines"`
	Costs    []CostTracking    `json:"costs"`
}

type UserTopFilter struct {
	CuisineTypes []c.CuisineChoices
	CostTypes    []int
}
