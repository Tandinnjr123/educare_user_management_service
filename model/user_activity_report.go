package model

import (
    "time"

    "github.com/google/uuid"
)

// UserActivityFilter narrows the user activity report. Activity is one of all, membership,
// coupon or login; FromDate/ToDate bound when that activity happened (ToDate inclusive).
type UserActivityFilter struct {
    Activity  string
    FromDate  *time.Time
    ToDate    *time.Time
    PackageID *uuid.UUID
    SearchBy  string
    Search    string
    Limit     int
    Offset    int
}

// UserActivityReportRow is one user with their latest membership, login and coupon activity.
type UserActivityReportRow struct {
    UserID               uuid.UUID  `json:"user_id"`
    Name                 *string    `json:"name"`
    Email                *string    `json:"email"`
    MobileNo             *string    `json:"mobile_no"`
    Status               *string    `json:"status"`
    PackageName          *string    `json:"package_name"`
    DurationType         *string    `json:"duration_type"`
    MembershipJoinDate   *time.Time `json:"membership_join_date"`
    MembershipEndDate    *time.Time `json:"membership_end_date"`
    MembershipUpdatedAt  *time.Time `json:"membership_updated_at"`
    LastLoginAt          *time.Time `json:"last_login_at"`
    CouponsRedeemed      int        `json:"coupons_redeemed"`
    LastCouponCode       *string    `json:"last_coupon_code"`
    LastCouponRedeemedAt *time.Time `json:"last_coupon_redeemed_at"`
}

type UserActivityReportResult struct {
    Users      []UserActivityReportRow `json:"users"`
    Total      int                     `json:"total"`
    Pagination *UserPagination         `json:"pagination"`
}
