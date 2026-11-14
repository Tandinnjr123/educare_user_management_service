package repositories

import (
    "user_management_service/model"
    "user_management_service/helpers"

    "fmt"
    "strings"
    "time"
    "errors"
    "context"

    "gorm.io/gorm"
	"github.com/google/uuid"
)

type UserRepository struct{
    DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{DB: db}
}

func (repo *UserRepository) CheckForExistingUser(field, value string) (*model.CommercialUser, error) {
    var user model.CommercialUser
    err := repo.DB.Where(fmt.Sprintf("%s = ? AND status != ?", field), value, "Deleted").First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find user with %v %v", field, value)
    }
    return &user, nil
}

func (repo *UserRepository) FetchUserByLoginID(field, value string) (*model.CommercialUser, error) {
    var user model.CommercialUser
    err := repo.DB.Where(fmt.Sprintf("%s = ?", field), value).First(&user).Error
    if err != nil {
        return nil, fmt.Errorf("failed to find user with %v %v", field, value)
    }
    return &user, nil
}


// CreateCommercialUser creates a new commercial user and their associated profile within a single transaction.
func (repo *UserRepository) CreateCommercialUser(signupInput *model.SignupInput) (*model.CommercialUser, *model.UserProfile, error) {
    var user *model.CommercialUser
    var userProfile *model.UserProfile
    var err error

    err = repo.DB.Transaction(func(tx *gorm.DB) error {
        identifier, err := helpers.GenerateRandomTokenString(6)
        if err != nil {
            return fmt.Errorf("failed to generate identifier: %v", err)
        }

        hashedPassword, err := helpers.EncryptPassword(signupInput.Password)
        if err != nil {
            return fmt.Errorf("failed to hash password: %v", err)
        }

        if signupInput.MobileNo != "" {
            user, _ = repo.FetchUserByLoginID("mobile_no", signupInput.MobileNo)
            if user == nil && signupInput.Email != "" {
                user, _ = repo.FetchUserByLoginID("email", signupInput.Email)
            }
        }

        if user != nil {
            updateData := map[string]interface{}{
                "mobile_no":       signupInput.MobileNo,
                "email":           signupInput.Email,
                "user_identifier": identifier,
                "password":        hashedPassword,
                "status":          "Active",
                "updated_at":      time.Now(),
            }
            if err := tx.Model(user).Updates(updateData).Error; err != nil {
                return fmt.Errorf("failed to update user data: %v", err)
            }
            return nil
        }

        newUser := model.CommercialUser{
            ID:             uuid.New(),
            MobileNo:       signupInput.MobileNo,
            Email:          signupInput.Email,
            UserIdentifier: identifier,
            Password:       hashedPassword,
            Status:         "Active",
            CreatedAt:      time.Now(),
            UpdatedAt:      time.Now(),
        }

        if err := tx.Create(&newUser).Error; err != nil {
            return fmt.Errorf("failed to insert user data: %v", err)
        }
        user = &newUser

        profileInput := model.UserProfileInput{
            UserId: newUser.ID,
            Gender: signupInput.Gender,
            Name:   signupInput.Name,
        }
        userProfile, err = repo.CreateUserProfile(tx, profileInput)
        if err != nil {
            return fmt.Errorf("failed to create user profile: %v", err)
        }

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    return user, userProfile, nil
}

func (repo *UserRepository) CreateUserProfile(tx *gorm.DB, inputData model.UserProfileInput) (*model.UserProfile, error) {
    userProfile := model.UserProfile{
        ID:             uuid.New(),
        Name:           inputData.Name,
        Gender:         inputData.Gender,
        ProfilePicture: inputData.ProfilePicture,
        UserId:         inputData.UserId,
    }

    if err := tx.Create(&userProfile).Error; err != nil {
        return nil, fmt.Errorf("failed to insert user profile: %v", err)
    }


    if err := tx.Save(&userProfile).Error; err != nil {
        return nil, fmt.Errorf("failed to update user profile with playlist ID: %v", err)
    }

    return &userProfile, nil
}

func (repo *UserRepository) FetchProfileByUserId(ctx context.Context, userId uuid.UUID) (*model.UserProfile, error) {
    var profile model.UserProfile

    db, err := helpers.GetGormDB()
    if err != nil {
        return nil, err
    }

    if err := db.WithContext(ctx).Where("user_id = ?", userId).First(&profile).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch profile for user id '%s': %w", userId, err)
    }
    return &profile, nil
}

func (repo *UserRepository) UpdateCommercialUser(userID uuid.UUID, signupInput *model.SignupInput) (*model.CommercialUser, *model.UserProfile, error) {
    var user model.CommercialUser
    var userProfile *model.UserProfile
    var err error

    err = repo.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.First(&user, "id = ?", userID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return fmt.Errorf("user with ID %s not found: %w", userID, err)
            }
            return fmt.Errorf("failed to fetch user for update: %w", err)
        }

        // Step 2: Prepare the update data for user.
        hashedPassword, err := helpers.EncryptPassword(signupInput.Password)
        if err != nil {
            return fmt.Errorf("failed to hash password: %w", err)
        }

        updateData := map[string]interface{}{
            "password":   hashedPassword,
            "updated_at": time.Now(),
        }
        if signupInput.MobileNo != "" {
            updateData["mobile_no"] = signupInput.MobileNo
        }
        if signupInput.Email != "" {
            updateData["email"] = signupInput.Email
        }

        if err := tx.Model(&model.CommercialUser{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
            return fmt.Errorf("failed to update user data: %w", err)
        }

        // --- Update the user profile as well ---
        var profile model.UserProfile
        if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
            return fmt.Errorf("failed to fetch user profile for update: %w", err)
        }

        profileUpdate := map[string]interface{}{
            "updated_at": time.Now(),
        }
        if signupInput.Name != "" {
            profileUpdate["name"] = signupInput.Name
        }
        if signupInput.Gender != "" {
            profileUpdate["gender"] = signupInput.Gender
        }

        if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(profileUpdate).Error; err != nil {
            return fmt.Errorf("failed to update user profile: %w", err)
        }

        userProfile = &profile

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    var updatedUser model.CommercialUser
    if err := repo.DB.First(&updatedUser, "id = ?", userID).Error; err != nil {
        return nil, nil, fmt.Errorf("failed to re-fetch updated user: %w", err)
    }

    // Re-fetch the updated profile
    userProfile, err = repo.FetchProfileByUserId(context.Background(), userID)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to re-fetch updated user profile: %w", err)
    }

    return &updatedUser, userProfile, nil
}

func (repo *UserRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) (*model.CommercialUser, error) {
	var user model.CommercialUser
	if err := repo.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found with ID: %s", userID)
	}

	updates := map[string]interface{}{
		"status": status,
		"updated_at":   time.Now(),
	}

	if err := repo.DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Fetch the updated question to return the new state.
	var updatedUser model.CommercialUser
	if err := repo.DB.WithContext(ctx).Where("id = ?", userID).First(&updatedUser).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch the updated USER: %w", err)
	}

	return &updatedUser, nil
}

func (repo *UserRepository) FetchCommercialUser(userID uuid.UUID) (*model.CommercialUser, error) {
    var user model.CommercialUser
    if err := repo.DB.Preload("UserProfile").First(&user, "id = ?", userID).Error; err != nil {
        return nil, fmt.Errorf("user not found: %v", err)
    }
    membership, err := repo.fetchLatestMembership(userID)
    if err != nil {
        return nil, err
    }
    user.Membership = membership
    return &user, nil
}

// fetchLatestMembership returns the user's most recently updated membership with its plan, or nil
// when the user has never had one or the subscription tables are not in this service's database.
func (repo *UserRepository) fetchLatestMembership(userID uuid.UUID) (*model.UserMembership, error) {
    if !relationsExist(repo.DB, "subscription.user_memberships", "subscription.membership_durations", "subscription.packages") {
        return nil, nil
    }
    var memberships []model.UserMembership
    err := repo.DB.Table("subscription.user_memberships um").
        Joins("LEFT JOIN subscription.membership_durations md ON md.id = um.membership_duration_id").
        Joins("LEFT JOIN subscription.packages pk ON pk.id = md.package_id").
        Select(`pk.name AS package_name, md.duration_type, md.number_of_days, um.membership_duration_id,
            um.membership_join_date, um.membership_end_date, um.membership_end_date > now() AS is_active`).
        Where("um.user_id = ?", userID).
        Order("um.updated_at DESC").
        Limit(1).
        Scan(&memberships).Error
    if err != nil {
        return nil, fmt.Errorf("failed to fetch membership: %v", err)
    }
    if len(memberships) == 0 {
        return nil, nil
    }
    return &memberships[0], nil
}

func (repo *UserRepository) FetchAllActiveUsers(limit, offset int) ([]model.CommercialUser, int, error) {
    var users []model.CommercialUser
    if err := repo.DB.Where("status = ?", "Active").
    Preload("UserProfile").
    Limit(limit).Offset(offset).
    Find(&users).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to fetch active users: %v", err)
    }

    var totalCount int64
    if err := repo.DB.Model(&model.CommercialUser{}).Where("status = ?", "Active").Count(&totalCount).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count active users: %v", err)
    }

    return users, int(totalCount), nil
}



func (repo *UserRepository) FetchAllCommercialUsers(limit, offset int, searchBy, search string) ([]model.CommercialUser, int, error) {
    var users []model.CommercialUser
    var totalCount int64

    filtered, err := applyUserSearch(repo.DB.Model(&model.CommercialUser{}), "auth.users", searchBy, search)
    if err != nil {
        return nil, 0, err
    }

    if err := filtered.Session(&gorm.Session{}).Count(&totalCount).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count users: %v", err)
    }

    db := filtered.Preload("UserProfile").Limit(limit).Offset(offset)
    if err := db.Find(&users).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to fetch users: %v", err) 
    }

    return users, int(totalCount), nil
}

func (repo *UserRepository) ResetPassword(userID uuid.UUID, newPassword string) error {
    hashedPassword, err := helpers.EncryptPassword(newPassword)
    if err != nil {
        return fmt.Errorf("failed to hash new password: %w", err)
    }

    // Update only the password and updated_at fields for the specified user.
    result := repo.DB.Model(&model.CommercialUser{}).
        Where("id = ?", userID).
        Updates(map[string]interface{}{
            "password":   hashedPassword,
            "updated_at": time.Now(),
        })

    if result.Error != nil {
        return fmt.Errorf("failed to update user password: %w", result.Error)
    }

    // Check if a user was actually updated.
    if result.RowsAffected == 0 {
        return fmt.Errorf("user with ID %s not found", userID)
    }

    return nil
}

func (repo *UserRepository) FetchNewRegister(from_date, to_date time.Time) ([]model.CommercialUser, error) {
    var users []model.CommercialUser
    if err := repo.DB.Where("created_at BETWEEN ? AND ?", from_date, to_date).Find(&users).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch new registrations: %v", err)
    }
    return users, nil
}

func (repo *UserRepository) GetCommercialUserTotals(fromDate, toDate *time.Time) (totalAll int, totalActive int, totalInActive,totalNew int, err error) {
    var totalAllCount int64
    var totalActiveCount int64
    var totalInActiveCount int64
    var totalNewCount int64

    query := repo.DB.Model(&model.CommercialUser{})
    if fromDate != nil && toDate != nil {
        query = query.Where("created_at BETWEEN ? AND ?", fromDate, toDate)
    }
    if err := query.Count(&totalAllCount).Error; err != nil {
        return 0, 0, 0, 0, fmt.Errorf("failed to count total users: %w", err)
    }

    activeQuery := repo.DB.Model(&model.UserActivity{}).
        Select("user_id").
        Group("user_id")
    if fromDate != nil && toDate != nil {
        activeQuery = activeQuery.Where("created_at BETWEEN ? AND ?", fromDate, toDate)
    }
    if err := activeQuery.Count(&totalActiveCount).Error; err != nil {
        return 0, 0, 0, 0, fmt.Errorf("failed to count active users: %w", err)
    }

    totalInActiveCount = totalAllCount - totalActiveCount
    if totalInActiveCount < 0 {
        totalInActiveCount = 0
    }

    if err := repo.DB.Model(&model.CommercialUser{}).Where("created_at BETWEEN ? AND ?", fromDate, toDate).Count(&totalNewCount).Error; err != nil {
        return 0, 0, 0, 0,fmt.Errorf("failed to count new registrations: %w", err)
    }

    return int(totalAllCount), int(totalActiveCount), int(totalNewCount), int(totalInActiveCount), nil
}

func (repo *UserRepository) GetUserActivity(offset, limit int) ([]model.UserActivity, error) {
    var userActivities []model.UserActivity
    
    var userIDs []uuid.UUID
    if err := repo.DB.Model(&model.UserActivity{}).
        Select("DISTINCT user_id").
        Order("user_id").
        Offset(offset).Limit(limit).
        Find(&userIDs).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch distinct user ids: %v", err)
    }

    if len(userIDs) == 0 {
        return []model.UserActivity{}, nil
    }
    if err := repo.DB.
        Preload("User").
        Preload("User.UserProfile").
        Where("user_id IN (?)", userIDs).
        Find(&userActivities).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch user activities: %v", err)
    }
    
    return userActivities, nil
}

// applyUserSearch narrows a users query by one field; usersTable is the table or alias holding
// auth.users. searchBy is matched against a fixed list so it can never reach the SQL as a column
// name. Name and email match any part of the value ignoring case; phone must match exactly.
func applyUserSearch(db *gorm.DB, usersTable, searchBy, search string) (*gorm.DB, error) {
    search = strings.TrimSpace(search)
    if search == "" {
        return db, nil
    }

    switch searchBy {
    case "name":
        return db.Where(usersTable+".id IN (?)", profilesByName(db, search)), nil
    case "email":
        return db.Where(usersTable+".email ILIKE ?", containsPattern(search)), nil
    case "mobile_no":
        return db.Where(usersTable+".mobile_no = ?", search), nil
    default:
        return nil, fmt.Errorf("invalid search_by %q: must be one of name, mobile_no, email", searchBy)
    }
}

func profilesByName(db *gorm.DB, search string) *gorm.DB {
    return db.Session(&gorm.Session{NewDB: true}).
        Model(&model.UserProfile{}).
        Select("user_id").
        Where("name ILIKE ?", containsPattern(search))
}

// containsPattern escapes LIKE wildcards so the value is matched literally.
func containsPattern(search string) string {
    escaped := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(search)
    return "%" + escaped + "%"
}
