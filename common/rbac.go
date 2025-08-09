package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"hpc-express-service/errors"
)

// Role represents a user role
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleSupervisor Role = "supervisor"
	RoleOperator   Role = "operator"
	RoleViewer     Role = "viewer"
)

// Permission represents a specific permission
type Permission string

const (
	// Cargo Manifest permissions
	PermissionCargoManifestView    Permission = "cargo_manifest:view"
	PermissionCargoManifestCreate  Permission = "cargo_manifest:create"
	PermissionCargoManifestUpdate  Permission = "cargo_manifest:update"
	PermissionCargoManifestDelete  Permission = "cargo_manifest:delete"
	PermissionCargoManifestConfirm Permission = "cargo_manifest:confirm"
	PermissionCargoManifestReject  Permission = "cargo_manifest:reject"
	PermissionCargoManifestPrint   Permission = "cargo_manifest:print"

	// Draft MAWB permissions
	PermissionDraftMAWBView    Permission = "draft_mawb:view"
	PermissionDraftMAWBCreate  Permission = "draft_mawb:create"
	PermissionDraftMAWBUpdate  Permission = "draft_mawb:update"
	PermissionDraftMAWBDelete  Permission = "draft_mawb:delete"
	PermissionDraftMAWBConfirm Permission = "draft_mawb:confirm"
	PermissionDraftMAWBReject  Permission = "draft_mawb:reject"
	PermissionDraftMAWBPrint   Permission = "draft_mawb:print"

	// MAWB Info permissions
	PermissionMAWBInfoView   Permission = "mawb_info:view"
	PermissionMAWBInfoCreate Permission = "mawb_info:create"
	PermissionMAWBInfoUpdate Permission = "mawb_info:update"
	PermissionMAWBInfoDelete Permission = "mawb_info:delete"

	// System permissions
	PermissionSystemAdmin Permission = "system:admin"
	PermissionAuditView   Permission = "audit:view"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// Admin has all permissions
		PermissionCargoManifestView, PermissionCargoManifestCreate, PermissionCargoManifestUpdate, PermissionCargoManifestDelete,
		PermissionCargoManifestConfirm, PermissionCargoManifestReject, PermissionCargoManifestPrint,
		PermissionDraftMAWBView, PermissionDraftMAWBCreate, PermissionDraftMAWBUpdate, PermissionDraftMAWBDelete,
		PermissionDraftMAWBConfirm, PermissionDraftMAWBReject, PermissionDraftMAWBPrint,
		PermissionMAWBInfoView, PermissionMAWBInfoCreate, PermissionMAWBInfoUpdate, PermissionMAWBInfoDelete,
		PermissionSystemAdmin, PermissionAuditView,
	},
	RoleSupervisor: {
		// Supervisor can view, create, update, and manage status
		PermissionCargoManifestView, PermissionCargoManifestCreate, PermissionCargoManifestUpdate,
		PermissionCargoManifestConfirm, PermissionCargoManifestReject, PermissionCargoManifestPrint,
		PermissionDraftMAWBView, PermissionDraftMAWBCreate, PermissionDraftMAWBUpdate,
		PermissionDraftMAWBConfirm, PermissionDraftMAWBReject, PermissionDraftMAWBPrint,
		PermissionMAWBInfoView, PermissionMAWBInfoCreate, PermissionMAWBInfoUpdate,
		PermissionAuditView,
	},
	RoleOperator: {
		// Operator can view, create, update, and print
		PermissionCargoManifestView, PermissionCargoManifestCreate, PermissionCargoManifestUpdate, PermissionCargoManifestPrint,
		PermissionDraftMAWBView, PermissionDraftMAWBCreate, PermissionDraftMAWBUpdate, PermissionDraftMAWBPrint,
		PermissionMAWBInfoView, PermissionMAWBInfoCreate, PermissionMAWBInfoUpdate,
	},
	RoleViewer: {
		// Viewer can only view and print
		PermissionCargoManifestView, PermissionCargoManifestPrint,
		PermissionDraftMAWBView, PermissionDraftMAWBPrint,
		PermissionMAWBInfoView,
	},
}

// User represents a user with roles and permissions
type User struct {
	ID          string       `json:"id"`
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Roles       []Role       `json:"roles"`
	Permissions []Permission `json:"permissions"`
	Active      bool         `json:"active"`
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role Role) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission Permission) bool {
	// Check direct permissions
	for _, userPermission := range u.Permissions {
		if userPermission == permission {
			return true
		}
	}

	// Check role-based permissions
	for _, userRole := range u.Roles {
		if rolePermissions, exists := RolePermissions[userRole]; exists {
			for _, rolePermission := range rolePermissions {
				if rolePermission == permission {
					return true
				}
			}
		}
	}

	return false
}

// HasAnyPermission checks if user has any of the specified permissions
func (u *User) HasAnyPermission(permissions ...Permission) bool {
	for _, permission := range permissions {
		if u.HasPermission(permission) {
			return true
		}
	}
	return false
}

// IsActive checks if user account is active
func (u *User) IsActive() bool {
	return u.Active
}

// RBACManager manages role-based access control
type RBACManager struct {
	enabled bool
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(enabled bool) *RBACManager {
	return &RBACManager{enabled: enabled}
}

// CheckPermission checks if a user has permission for an operation
func (rbac *RBACManager) CheckPermission(user *User, permission Permission) error {
	if !rbac.enabled {
		return nil // RBAC disabled, allow all operations
	}

	if user == nil {
		return errors.NewBusinessRuleError("authentication", "user not authenticated")
	}

	if !user.IsActive() {
		return errors.NewBusinessRuleError("authorization", "user account is inactive")
	}

	if !user.HasPermission(permission) {
		return errors.NewBusinessRuleError("authorization", fmt.Sprintf("insufficient permissions: %s required", permission))
	}

	return nil
}

// CheckAnyPermission checks if a user has any of the specified permissions
func (rbac *RBACManager) CheckAnyPermission(user *User, permissions ...Permission) error {
	if !rbac.enabled {
		return nil // RBAC disabled, allow all operations
	}

	if user == nil {
		return errors.NewBusinessRuleError("authentication", "user not authenticated")
	}

	if !user.IsActive() {
		return errors.NewBusinessRuleError("authorization", "user account is inactive")
	}

	if !user.HasAnyPermission(permissions...) {
		permissionNames := make([]string, len(permissions))
		for i, perm := range permissions {
			permissionNames[i] = string(perm)
		}
		return errors.NewBusinessRuleError("authorization", fmt.Sprintf("insufficient permissions: one of [%s] required", strings.Join(permissionNames, ", ")))
	}

	return nil
}

// GetUserFromContext extracts user information from request context
func (rbac *RBACManager) GetUserFromContext(ctx context.Context) (*User, error) {
	// This would typically extract user info from JWT claims or session
	// For now, we'll create a mock implementation
	userID := ctx.Value("userID")
	if userID == nil {
		return nil, errors.NewBusinessRuleError("authentication", "user not found in context")
	}

	// In a real implementation, you would fetch user details from database
	// For now, return a mock user with operator role
	return &User{
		ID:       userID.(string),
		Username: "operator",
		Email:    "operator@example.com",
		Roles:    []Role{RoleOperator},
		Active:   true,
	}, nil
}

// RequirePermission creates middleware that requires specific permission
func (rbac *RBACManager) RequirePermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := rbac.GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if err := rbac.CheckPermission(user, permission); err != nil {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Add user to context for downstream handlers
			ctx := context.WithValue(r.Context(), "currentUser", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func (rbac *RBACManager) RequireAnyPermission(permissions ...Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := rbac.GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if err := rbac.CheckAnyPermission(user, permissions...); err != nil {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Add user to context for downstream handlers
			ctx := context.WithValue(r.Context(), "currentUser", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates middleware that requires specific role
func (rbac *RBACManager) RequireRole(role Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := rbac.GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !user.HasRole(role) {
				http.Error(w, "Insufficient role", http.StatusForbidden)
				return
			}

			// Add user to context for downstream handlers
			ctx := context.WithValue(r.Context(), "currentUser", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Global RBAC manager instance
var GlobalRBACManager = NewRBACManager(true)

// Helper functions for permission checks

// CheckCargoManifestPermission checks cargo manifest permissions
func CheckCargoManifestPermission(ctx context.Context, operation string) error {
	user, err := GlobalRBACManager.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	var permission Permission
	switch operation {
	case "view":
		permission = PermissionCargoManifestView
	case "create", "update":
		permission = PermissionCargoManifestUpdate
	case "confirm":
		permission = PermissionCargoManifestConfirm
	case "reject":
		permission = PermissionCargoManifestReject
	case "print":
		permission = PermissionCargoManifestPrint
	default:
		return errors.NewBusinessRuleError("authorization", "unknown operation: "+operation)
	}

	return GlobalRBACManager.CheckPermission(user, permission)
}

// CheckDraftMAWBPermission checks draft MAWB permissions
func CheckDraftMAWBPermission(ctx context.Context, operation string) error {
	user, err := GlobalRBACManager.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	var permission Permission
	switch operation {
	case "view":
		permission = PermissionDraftMAWBView
	case "create", "update":
		permission = PermissionDraftMAWBUpdate
	case "confirm":
		permission = PermissionDraftMAWBConfirm
	case "reject":
		permission = PermissionDraftMAWBReject
	case "print":
		permission = PermissionDraftMAWBPrint
	default:
		return errors.NewBusinessRuleError("authorization", "unknown operation: "+operation)
	}

	return GlobalRBACManager.CheckPermission(user, permission)
}
