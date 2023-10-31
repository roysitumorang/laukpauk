package model

import (
	"net"
	"time"

	regionModel "github.com/roysitumorang/laukpauk/modules/region/model"
	roleModel "github.com/roysitumorang/laukpauk/modules/role/model"
)

const (
	StatusHold int = iota
	StatusActive
	StatusSuspended = -1
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
		RegistrationIP       net.IP             `json:"registration_ip"`
		Gender               *string            `json:"gender"`
		DateOfBirth          *time.Time         `json:"date_of_birth"`
		Avatar               *string            `json:"avatar"`
		Thumbnails           *string            `json:"thumbnails"`
		BusinessDays         *BusinessDays      `json:"business_days,omitempty"`
		BusinessOpeningHour  *int               `json:"business_opening_hour"`
		BusinessClosingHour  *int               `json:"business_closing_hour"`
		DeliveryHours        []int              `json:"delivery_hours"`
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

	BusinessDays struct {
		Sunday    bool `json:"sunday"`
		Monday    bool `json:"monday"`
		Tuesday   bool `json:"tuesday"`
		Wednesday bool `json:"wednesday"`
		Thursday  bool `json:"thursday"`
		Friday    bool `json:"friday"`
		Saturday  bool `json:"saturday"`
	}

	UserFilter struct {
		UserIDs,
		RoleIDs []int64
		MobilePhones []string
		Status       []int
	}
)
