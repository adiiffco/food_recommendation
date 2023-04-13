package utils

import "encoding/json"

func ConvertToObj(source, dest interface{}) error {
	bs, err := json.Marshal(source)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, &dest)
	if err != nil {
		return err
	}
	return nil
}

func GetKeyForRating(rating float64) string {
	if rating >= 4.5 {
		return "4.5_5"
	}
	if rating >= 4 {
		return "4_4.5"
	}
	return "0_4"
}
