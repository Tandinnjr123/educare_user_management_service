package repositories

import (
    "fmt"
    "time"

    "user_management_service/model"

    "gorm.io/gorm"
)

// FetchUserActivityReport lists commercial users with their current membership, last login and
// redeemed coupons. Reads the subscription, coupon and auth schemas of the shared database.
func (repo *UserRepository) FetchUserActivityReport(filter model.UserActivityFilter) ([]model.UserActivityReportRow, int, error) {
    query := repo.DB.Table("auth.users AS u").
        Joins("LEFT JOIN user_data.user_profiles p ON p.user_id = u.id").
        Joins(`LEFT JOIN LATERAL (
            SELECT um.membership_join_date, um.membership_end_date, um.updated_at, md.duration_type, pk.id AS package_id, pk.name AS package_name
            FROM subscription.user_memberships um
            LEFT JOIN subscription.membership_durations md ON md.id = um.membership_duration_id
            LEFT JOIN subscription.packages pk ON pk.id = md.package_id
            WHERE um.user_id = u.id
            ORDER BY um.updated_at DESC
            LIMIT 1) m ON true`).
        Joins("LEFT JOIN LATERAL (SELECT MAX(t.created_at) AS last_login_at FROM auth.access_tokens t WHERE t.user_id = u.id) l ON true").
        Joins(`LEFT JOIN LATERAL (
            SELECT COUNT(*) AS coupons_redeemed, MAX(c.redeemed_at) AS last_coupon_redeemed_at,
                   (ARRAY_AGG(c.code ORDER BY c.redeemed_at DESC))[1] AS last_coupon_code
            FROM coupon.coupons c
            WHERE c.user_id = u.id AND c.redeemed) cp ON true`)

    query, err := applyActivityFilter(query, filter)
    if err != nil {
        return nil, 0, err
    }
    if filter.PackageID != nil {
        query = query.Where("m.package_id = ?", *filter.PackageID)
    }
    if query, err = applyUserSearch(query, "u", filter.SearchBy, filter.Search); err != nil {
        return nil, 0, err
    }

    var total int64
    if err := repo.DB.Table("(?) AS filtered", query.Session(&gorm.Session{}).Select("u.id")).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count user activities: %v", err)
    }

    var rows []model.UserActivityReportRow
    err = query.Select(`u.id AS user_id, p.name, u.email, u.mobile_no, u.status,
            m.package_name, m.duration_type, m.membership_join_date, m.membership_end_date, m.updated_at AS membership_updated_at,
            l.last_login_at, COALESCE(cp.coupons_redeemed, 0) AS coupons_redeemed, cp.last_coupon_code, cp.last_coupon_redeemed_at`).
        Order("GREATEST(l.last_login_at, m.updated_at, cp.last_coupon_redeemed_at) DESC NULLS LAST, u.created_at DESC").
        Limit(filter.Limit).Offset(filter.Offset).
        Scan(&rows).Error
    if err != nil {
        return nil, 0, fmt.Errorf("failed to fetch user activities: %v", err)
    }
    return rows, int(total), nil
}

// applyActivityFilter keeps users who did the chosen activity, within the date range when one is
// given. Without a date range, "all" keeps every user.
func applyActivityFilter(query *gorm.DB, filter model.UserActivityFilter) (*gorm.DB, error) {
    from, to := activityWindow(filter.FromDate, filter.ToDate)
    membership := gorm.Expr("m.updated_at >= ? AND m.updated_at < ?", from, to)
    login := gorm.Expr("EXISTS (SELECT 1 FROM auth.access_tokens t WHERE t.user_id = u.id AND t.created_at >= ? AND t.created_at < ?)", from, to)
    coupon := gorm.Expr("EXISTS (SELECT 1 FROM coupon.coupons c WHERE c.user_id = u.id AND c.redeemed AND c.redeemed_at >= ? AND c.redeemed_at < ?)", from, to)

    switch filter.Activity {
    case "", "all":
        if filter.FromDate == nil && filter.ToDate == nil {
            return query, nil
        }
        return query.Where("(?) OR (?) OR (?)", membership, login, coupon), nil
    case "membership":
        return query.Where(membership), nil
    case "login":
        return query.Where(login), nil
    case "coupon":
        return query.Where(coupon), nil
    default:
        return nil, fmt.Errorf("invalid activity %q: must be one of all, membership, coupon, login", filter.Activity)
    }
}

// activityWindow turns an optional inclusive date range into a half-open time range.
func activityWindow(from, to *time.Time) (time.Time, time.Time) {
    start := time.Time{}
    end := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
    if from != nil {
        start = *from
    }
    if to != nil {
        end = to.AddDate(0, 0, 1)
    }
    return start, end
}
