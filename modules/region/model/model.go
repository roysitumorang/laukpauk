package model

type (
	Region struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	Village struct {
		Region
		SubdistrictID int64 `json:"subdistrict_id"`
	}
)
