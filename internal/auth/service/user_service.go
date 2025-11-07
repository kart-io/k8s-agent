package service

import (
	stderrors "errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/auth/crypto"
	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	"github.com/kart-io/logger/core"
)

// UserService handles user management business logic.
type UserService struct {
	db     *storage.MySQLDB
	logger core.Logger
}

// NewUserService creates a new user service.
func NewUserService(db *storage.MySQLDB, logger core.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger.With("component", "user-service"),
	}
}

// List retrieves users with pagination and filtering.
func (s *UserService) List(params types.PaginationParams, statusFilter *int) (*types.PaginatedResponse, error) {
	s.logger.Infow("List users request",
		"page", params.Page,
		"page_size", params.PageSize,
		"status_filter", statusFilter,
	)

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
		s.logger.Errorw("List users failed: count error", "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Query with pagination and sorting
	var modelUsers []model.User
	orderClause := params.Sort + " " + params.Order
	if err := query.Order(orderClause).Limit(params.PageSize).Offset(offset).Find(&modelUsers).Error; err != nil {
		s.logger.Errorw("List users failed: query error", "error", err)
		return nil, errors.NewDatabaseError(err)
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

	s.logger.Infow("List users successful",
		"total", total,
		"page", params.Page,
		"page_size", params.PageSize,
		"returned", len(users),
	)

	return &types.PaginatedResponse{
		Items:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(id string) (*types.User, error) {
	s.logger.Infow("Get user by ID request", "user_id", id)

	var user model.User
	err := s.db.DB.Where("id = ?", id).First(&user).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Get user failed: user not found", "user_id", id)
		return nil, errors.ErrNotFound.WithMessage("user not found").KV("user_id", id)
	}
	if err != nil {
		s.logger.Errorw("Get user failed: database error", "user_id", id, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.Infow("Get user successful", "user_id", id, "username", user.Username)

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

// Create creates a new user.
func (s *UserService) Create(req *types.UserCreateRequest) (*types.User, error) {
	s.logger.Infow("Create user request", "username", req.Username, "email", req.Email)

	// Hash password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorw("Create user failed: password hashing error", "username", req.Username, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to hash password")
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
			return errors.NewDatabaseError(err)
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
					return errors.NewDatabaseError(err)
				}
			}
		} else {
			// Assign default 'user' role
			var defaultRole model.Role
			if err := tx.Where("code = ?", "user").First(&defaultRole).Error; err != nil {
				return errors.ErrNotFound.WithMessage("default role not found")
			}
			userRole := model.UserRole{
				UserID: userID,
				RoleID: defaultRole.ID,
			}
			if err := tx.Create(&userRole).Error; err != nil {
				return errors.NewDatabaseError(err)
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Errorw("Create user failed: transaction error",
			"username", req.Username,
			"error", err,
		)
		return nil, err
	}

	s.logger.Infow("Create user successful",
		"user_id", userID,
		"username", req.Username,
		"roles_count", len(req.RoleIDs),
	)

	return s.GetByID(userID)
}

// Update updates user information.
func (s *UserService) Update(id string, req *types.UserUpdateRequest) error {
	s.logger.Infow("Update user request", "user_id", id)

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
			return errors.NewDatabaseError(result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.ErrNotFound.WithMessage("user not found").KV("user_id", id)
		}

		// Update roles if provided
		if req.RoleIDs != nil {
			// Use GORM Association to replace roles
			var user model.User
			user.ID = id

			// Load roles by IDs
			var roles []model.Role
			if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
				return errors.NewDatabaseError(err)
			}

			// Replace associations
			if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
				return errors.NewDatabaseError(err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Errorw("Update user failed", "user_id", id, "error", err)
		return err
	}

	s.logger.Infow("Update user successful", "user_id", id)
	return nil
}

// Delete soft deletes a user (sets status to 0).
func (s *UserService) Delete(id string) error {
	s.logger.Infow("Delete user request", "user_id", id)

	result := s.db.DB.Model(&model.User{}).Where("id = ?", id).Update("status", 0)
	if result.Error != nil {
		s.logger.Errorw("Delete user failed: database error", "user_id", id, "error", result.Error)
		return errors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Warnw("Delete user failed: user not found", "user_id", id)
		return errors.ErrNotFound.WithMessage("user not found").KV("user_id", id)
	}

	s.logger.Infow("Delete user successful", "user_id", id)
	return nil
}

// AssignRoles assigns roles to a user.
func (s *UserService) AssignRoles(userID string, roleIDs []string) error {
	s.logger.Infow("Assign roles request", "user_id", userID, "roles_count", len(roleIDs))

	var user model.User
	user.ID = userID

	// Load roles by IDs
	var roles []model.Role
	if err := s.db.DB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		s.logger.Errorw("Assign roles failed: database error", "user_id", userID, "error", err)
		return errors.NewDatabaseError(err)
	}

	// Use GORM Association to replace roles
	if err := s.db.DB.Model(&user).Association("Roles").Replace(roles); err != nil {
		s.logger.Errorw("Assign roles failed: association error", "user_id", userID, "error", err)
		return errors.NewDatabaseError(err)
	}

	s.logger.Infow("Assign roles successful", "user_id", userID, "roles_count", len(roles))
	return nil
}
