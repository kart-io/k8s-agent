package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/auth-service/internal/model"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	"github.com/kart-io/k8s-agent/auth-service/pkg/crypto"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
	"gorm.io/gorm"
)

// UserService handles user management business logic
type UserService struct {
	db *storage.PostgresDB
}

// NewUserService creates a new user service
func NewUserService(db *storage.PostgresDB) *UserService {
	return &UserService{db: db}
}

// List retrieves users with pagination and filtering
func (s *UserService) List(params types.PaginationParams, statusFilter *int) (*types.PaginatedResponse, error) {
	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.Sort == "" {
		params.Sort = "created_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}

	offset := (params.Page - 1) * params.PageSize

	// Build GORM query with filters
	query := s.db.DB.Model(&model.User{})

	if statusFilter != nil {
		query = query.Where("status = ?", *statusFilter)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Query with pagination and sorting
	var modelUsers []model.User
	orderClause := fmt.Sprintf("%s %s", params.Sort, params.Order)
	if err := query.Order(orderClause).Limit(params.PageSize).Offset(offset).Find(&modelUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	// Convert model.User to types.User
	users := make([]types.User, len(modelUsers))
	for i, u := range modelUsers {
		users[i] = types.User{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			RealName:  u.RealName,
			Phone:     u.Phone,
			Avatar:    u.Avatar,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &types.PaginatedResponse{
		Items:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id string) (*types.User, error) {
	var user model.User
	err := s.db.DB.Where("id = ?", id).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &types.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Create creates a new user
func (s *UserService) Create(req *types.UserCreateRequest) (*types.User, error) {
	// Hash password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID := uuid.New().String()

	// Create user model
	user := &model.User{
		ID:       userID,
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		RealName: req.RealName,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   1,
	}

	// Begin transaction
	err = s.db.DB.Transaction(func(tx *gorm.DB) error {
		// Insert user
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		// Assign roles
		if len(req.RoleIDs) > 0 {
			// Assign provided roles
			for _, roleID := range req.RoleIDs {
				userRole := model.UserRole{
					UserID: userID,
					RoleID: roleID,
				}
				if err := tx.Create(&userRole).Error; err != nil {
					return fmt.Errorf("failed to assign role: %w", err)
				}
			}
		} else {
			// Assign default 'user' role
			var defaultRole model.Role
			if err := tx.Where("code = ?", "user").First(&defaultRole).Error; err != nil {
				return fmt.Errorf("failed to find default role: %w", err)
			}
			userRole := model.UserRole{
				UserID: userID,
				RoleID: defaultRole.ID,
			}
			if err := tx.Create(&userRole).Error; err != nil {
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetByID(userID)
}

// Update updates user information
func (s *UserService) Update(id string, req *types.UserUpdateRequest) error {
	// Build update data map for non-empty fields
	updateData := make(map[string]interface{})

	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.RealName != "" {
		updateData["real_name"] = req.RealName
	}
	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}
	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}

	// Begin transaction
	err := s.db.DB.Transaction(func(tx *gorm.DB) error {
		// Update user
		result := tx.Model(&model.User{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			return fmt.Errorf("failed to update user: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		// Update roles if provided
		if req.RoleIDs != nil {
			// Use GORM Association to replace roles
			var user model.User
			user.ID = id

			// Load roles by IDs
			var roles []model.Role
			if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
				return fmt.Errorf("failed to find roles: %w", err)
			}

			// Replace associations
			if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
				return fmt.Errorf("failed to update roles: %w", err)
			}
		}

		return nil
	})

	return err
}

// Delete soft deletes a user (sets status to 0)
func (s *UserService) Delete(id string) error {
	result := s.db.DB.Model(&model.User{}).Where("id = ?", id).Update("status", 0)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// AssignRoles assigns roles to a user
func (s *UserService) AssignRoles(userID string, roleIDs []string) error {
	var user model.User
	user.ID = userID

	// Load roles by IDs
	var roles []model.Role
	if err := s.db.DB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return fmt.Errorf("failed to find roles: %w", err)
	}

	// Use GORM Association to replace roles
	if err := s.db.DB.Model(&user).Association("Roles").Replace(roles); err != nil {
		return fmt.Errorf("failed to assign roles: %w", err)
	}

	return nil
}
