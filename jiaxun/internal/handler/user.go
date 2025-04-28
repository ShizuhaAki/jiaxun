package handler

import (
	"errors"
	"net/http"
	"strconv"

	"jiaxun/internal/middleware"
	"jiaxun/internal/model"
	"jiaxun/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler and registers routes
func NewUserHandler(r *gin.Engine, userService *service.UserService) *UserHandler {
	handler := &UserHandler{
		userService: userService,
	}

	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", handler.Login)
	}

	// All protected routes require auth middleware globally
	users := r.Group("/api/users")
	users.Use(middleware.AuthMiddleware())
	{
		// Routes for all authenticated users
		users.GET("/me", handler.GetCurrentUser)
		users.GET("/:id", handler.GetUser)

		// These routes are accessible only to self and to teachers
		userOwnerGroup := users.Group("/:id")
		userOwnerGroup.Use(middleware.CanModifyUser())
		{
			userOwnerGroup.PUT("", handler.UpdateUser)
			userOwnerGroup.DELETE("", handler.DeleteUser)
		}

		// Admin (teacher)-only routes
		adminGroup := users.Group("")
		adminGroup.Use(middleware.TeacherRequired())
		{
			adminGroup.GET("", handler.ListUsers)
			adminGroup.POST("", handler.Register)
		}

	}

	return handler
}

// @Summary Register a new user
// @Description Creates a new user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param body body object true "User information"
// @Success 201 {object} object{user=model.User} "Created user"
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 409 {object} object{error=string} "User already exists"
// @Failure 500 {object} object{error=string} "Server error"
// @id Register
// @Router /users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &model.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
		FullName: request.FullName,
	}

	if err := h.userService.Create(user); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// @Summary User login
// @Description Authenticates a user and returns a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{username=string,password=string} true "Login credentials"
// @Success 200 {object} object{token=string,user=object{id=integer,username=string,email=string,fullName=string,role=string}} "Login successful"
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Invalid credentials"
// @Failure 500 {object} object{error=string} "Server error"
// @id Login
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Authenticate(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"fullName": user.FullName,
			"role":     user.Role,
		},
	})
}

// @Summary Get user by ID
// @Description Retrieves a user's profile by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path integer true "User ID"
// @Success 200 {object} object{user=model.User} "User found"
// @Failure 400 {object} object{error=string} "Invalid user ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{id} [get]
// @id GetUser
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// @Summary Get current user
// @Description Retrieves the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} object{user=model.User} "Current user"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/me [get]
// @id GetCurrentUser
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID") // This will always exist due to auth middleware

	user, err := h.userService.GetByID(userID.(uint))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// @Summary Update user
// @Description Updates a user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Param id path integer true "User ID"
// @Param body body object{email=string,password=string,full_name=string} false "Fields to update"
// @Success 200 {object} object{user=model.User} "Updated user"
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{id} [put]
// @id UpdateUser
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Get target user ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing user
	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Update fields if provided
	if request.Email != "" {
		user.Email = request.Email
	}
	if request.Password != "" {
		user.Password = request.Password
	}
	if request.FullName != "" {
		user.FullName = request.FullName
	}

	if err := h.userService.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// @Summary Delete user
// @Description Removes a user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path integer true "User ID"
// @Success 200 {object} object{message=string} "User deleted successfully"
// @Failure 400 {object} object{error=string} "Invalid user ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users/{id} [delete]
// @id DeleteUser
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Get target user ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userService.Delete(uint(id)); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// @Summary List users
// @Description Returns a paginated list of users (teachers only)
// @Tags users
// @Accept json
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param page_size query integer false "Page size (default: 10, max: 100)"
// @Success 200 {object} array{User} "List of users"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Server error"
// @Router /users [get]
// @id ListUsers
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	// Remove passwords from response
	for _, user := range users {
		user.Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
