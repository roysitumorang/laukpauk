package model

import (
	"time"
)

type (
	Banner struct {
		ID         int64     `json:"id"`
		ParentID   *int64    `json:"parent_id"`
		UserID     *int64    `json:"user_id"`
		Published  bool      `json:"published"`
		File       string    `json:"file"`
		Thumbnails []Banner  `json:"thumbnails,omitempty"`
		CreatedBy  int64     `json:"created_by"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedBy  int64     `json:"updated_by"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
)
