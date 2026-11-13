package schema

import "github.com/graphql-go/graphql"

var CommercialUserActivity = graphql.NewObject(graphql.ObjectConfig{
	Name: "CommercialUserActivity",
	Fields: graphql.Fields{
		"user_id":                 &graphql.Field{Type: graphql.String},
		"name":                    &graphql.Field{Type: graphql.String},
		"email":                   &graphql.Field{Type: graphql.String},
		"mobile_no":               &graphql.Field{Type: graphql.String},
		"status":                  &graphql.Field{Type: graphql.String},
		"package_name":            &graphql.Field{Type: graphql.String},
		"duration_type":           &graphql.Field{Type: graphql.String},
		"membership_join_date":    &graphql.Field{Type: graphql.DateTime},
		"membership_end_date":     &graphql.Field{Type: graphql.DateTime},
		"membership_updated_at":   &graphql.Field{Type: graphql.DateTime},
		"last_login_at":           &graphql.Field{Type: graphql.DateTime},
		"coupons_redeemed":        &graphql.Field{Type: graphql.Int},
		"last_coupon_code":        &graphql.Field{Type: graphql.String},
		"last_coupon_redeemed_at": &graphql.Field{Type: graphql.DateTime},
	},
})

var CommercialUserActivityPagination = graphql.NewObject(graphql.ObjectConfig{
	Name: "CommercialUserActivityPagination",
	Fields: graphql.Fields{
		"current_page": &graphql.Field{Type: graphql.Int},
		"total_page":   &graphql.Field{Type: graphql.Int},
		"limit":        &graphql.Field{Type: graphql.Int},
	},
})

var CommercialUserActivitiesResult = graphql.NewObject(graphql.ObjectConfig{
	Name: "CommercialUserActivitiesResult",
	Fields: graphql.Fields{
		"users":      &graphql.Field{Type: graphql.NewList(CommercialUserActivity)},
		"total":      &graphql.Field{Type: graphql.Int},
		"pagination": &graphql.Field{Type: CommercialUserActivityPagination},
	},
})

var CommercialUserActivitiesResponse = graphql.NewObject(graphql.ObjectConfig{
	Name: "CommercialUserActivitiesResponse",
	Fields: graphql.Fields{
		"data":  &graphql.Field{Type: CommercialUserActivitiesResult},
		"error": &graphql.Field{Type: CommercialAuthError},
	},
})
