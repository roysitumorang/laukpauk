package model

import (
	"time"

	regionModel "github.com/roysitumorang/laukpauk/modules/region/model"
	roleModel "github.com/roysitumorang/laukpauk/modules/role/model"
)

type (
	User struct {
		ID                   int64              `json:"id"`
		Role                 roleModel.Role     `json:"role"`
		ApiKey               *string            `json:"api_key"`
		MerchantNote         *string            `json:"merchant_note"`
		MinimumPurchase      int                `json:"minimum_purchase"`
		AdminFee             int                `json:"admin_fee"`
		AccumulationDivisor  int                `json:"accumulation_divisor"`
		Name                 string             `json:"name"`
		Email                *string            `json:"email"`
		Password             string             `json:"-"`
		Address              string             `json:"address"`
		Village              regionModel.Region `json:"village"`
		Subdistrict          regionModel.Region `json:"subdistrict"`
		City                 regionModel.Region `json:"city"`
		Province             regionModel.Region `json:"province"`
		MobilePhone          string             `json:"mobile_phone"`
		DeviceToken          *string            `json:"device_token"`
		Status               int                `json:"status"`
		ActivatedAt          *time.Time         `json:"activated_at"`
		ActivationToken      *string            `json:"activation_token"`
		PasswordResetToken   *string            `json:"password_reset_token"`
		Deposit              float64            `json:"deposit"`
		Company              *string            `json:"company"`
		RegistrationIP       string             `json:"registration_ip"`
		Gender               *string            `json:"gender"`
		DateOfBirth          *time.Time         `json:"date_of_birth"`
		Avatar               *string            `json:"avatar"`
		Thumbnails           *string            `json:"thumbnails"`
		OpenOnSunday         int                `json:"open_on_sunday"`
		OpenOnMonday         int                `json:"open_on_monday"`
		OpenOnTuesday        int                `json:"open_on_tuesday"`
		OpenOnWednesday      int                `json:"open_on_wednesday"`
		OpenOnThursday       int                `json:"open_on_thursday"`
		OpenOnFriday         int                `json:"open_on_friday"`
		OpenOnSaturday       int                `json:"open_on_saturday"`
		BusinessOpeningHour  *string            `json:"business_opening_hour"`
		BusinessClosingHour  *string            `json:"business_closing_hour"`
		DeliveryHours        *string            `json:"delivery_hours"`
		Latitude             *float64           `json:"latitude"`
		Longitude            *float64           `json:"longitude"`
		Keywords             *string            `json:"keywords"`
		DeliveryMaxDistance  int                `json:"delivery_max_distance"`
		DeliveryFreeDistance int                `json:"delivery_free_distance"`
		DeliveryRate         int                `json:"delivery_rate"`
		CreatedBy            *int64             `json:"created_by"`
		CreatedAt            *time.Time         `json:"created_at"`
		UpdatedBy            *int64             `json:"updated_by"`
		UpdatedAt            *time.Time         `json:"updated_at"`
	}

	UserFilter struct {
		UserIDs,
		RoleIDs []int64
		MobilePhones []string
		Status       []int
	}
)
