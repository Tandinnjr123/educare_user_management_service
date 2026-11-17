package helpers

import "cms_user_management_service/model"

func FormatError(err error) *model.GenericUserResponse {
	return &model.GenericUserResponse{
		Data: nil,
		Error: &model.UserError{
			Message: err.Error(),
		},
	}
}
