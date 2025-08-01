package controller

import (
	"biostat/auth"
	"biostat/config"
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"context"
	"errors"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	patientService      service.PatientService
	roleService         service.RoleService
	userService         service.UserService
	emailService        service.EmailService
	authService         auth.AuthService
	permissionService   service.PermissionService
	subscriptionService service.SubscriptionService
}

func NewUserController(patientService service.PatientService, roleService service.RoleService,
	userService service.UserService, emailService service.EmailService, authService auth.AuthService,
	permissionService service.PermissionService, subscriptionService service.SubscriptionService) *UserController {
	return &UserController{patientService: patientService, roleService: roleService,
		userService: userService, emailService: emailService, authService: authService, permissionService: permissionService, subscriptionService: subscriptionService}
}

func (uc *UserController) RegisterUser(c *gin.Context) {
	var user models.SystemUser_
	if err := c.ShouldBindJSON(&user); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to bind user object", nil, err)
		return
	}
	var rawPassword string
	var hashedPassword []byte
	var err error

	if user.RoleName == "doctor" || user.RoleName == "caregiver" || user.RoleName == "nurse" || user.RoleName == "pharmacist" || (user.RoleName == "patient" && user.AuthType == "GMAIL") {
		rawPassword = utils.GenerateRandomPassword()
	} else {
		rawPassword = user.Password
	}

	hashedPassword, err = bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to hash password", nil, err)
		return
	}
	roleMaster, err := uc.roleService.GetRoleIdByRoleName(user.RoleName)
	if err != nil {
		log.Println("Error fetching role from role master:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Role not found", nil, err)
		return
	} else {
		log.Printf("Role Id: %d, Role Name: %s\n", roleMaster.RoleId, roleMaster.RoleName)
	}
	user.RoleName = roleMaster.RoleName
	user.RoleId = roleMaster.RoleId
	user.Email = strings.ToLower(user.Email)
	user.Username = uc.userService.GenerateUniqueUsername(user.FirstName, user.LastName)

	keyCloakUser := user
	keyCloakUser.Password = rawPassword
	//Add User in Keycloak
	keyCloakID, _, err := uc.authService.CreateUserInKeycloak(keyCloakUser)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, keyCloakID, nil, err)
		return
	}
	user.Password = string(hashedPassword)
	user.AuthUserId = keyCloakID

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in RegisterUser:", r)
			debug.PrintStack()
		}
	}()
	systemUser, err := uc.userService.CreateSystemUser(tx, user)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to register user", nil, err)
		return
	}
	mappingError := uc.roleService.AddSystemUserMapping(tx, nil, systemUser, &systemUser, roleMaster.RoleId, roleMaster.RoleName, nil, nil, nil)
	if mappingError != nil {
		log.Println("Error while adding user mapping", mappingError)
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User mapping not added", nil, mappingError)
		return
	}
	log.Println("System User before commiting:", mappingError)
	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to commit transaction", nil, err)
		return
	}
	err = uc.emailService.SendLoginCredentials(systemUser, nil, nil, "")
	if err != nil {
		log.Println("Error sending email:", err)
	}
	response := utils.MapUserToRoleSchema(systemUser, roleMaster.RoleName)
	subscribeErr := uc.subscriptionService.SubscribeDefaultPlan(systemUser.UserId, roleMaster.RoleName, user.LastName)
	if subscribeErr != nil {
		log.Println("Default subscribe pkan failed ", subscribeErr)
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User registered successfully", response, nil, nil)
	return
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var user models.SystemUser_
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		// Type     string  `json:"type" binding:"omitempty"`
		Password string  `json:"password" binding:"required"`
		LoginAs  *string `json:"login_as"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Missing required fields.", nil, err)
		return
	}

	loginInfo, err := uc.userService.GetUserInfoByIdentifier(strings.ToLower(input.Username))
	if err != nil {
		log.Println("User not found with this username in database : ", input.Username)
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User not found", nil, nil)
	}
	if !ComparePasswords(loginInfo.Password, input.Password) {
		log.Println("Password not match with hashpassword ")
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid user credentials!", nil, nil)
		return
	}
	client := config.Client
	ctx := context.Background()
	token, err := client.Login(ctx, config.KeycloakClientID, config.KeycloakClientSecret, config.KeycloakRealm, loginInfo.Username, input.Password)
	if err != nil {
		log.Println("Error logging in to Keycloak:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid user credentials!", nil, err)
		return
	}
	_, claims, err := client.DecodeAccessToken(ctx, token.AccessToken, config.KeycloakRealm)
	if err != nil {
		log.Println("Error decoding access token:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token", nil, err)
		c.Abort()
		return
	}
	if claims == nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token claims", nil, err)
		c.Abort()
		return
	}

	sub, ok := (*claims)["sub"]
	if !ok {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token claims", nil, err)
		c.Abort()
		return
	}
	realmAccessRaw, ok := (*claims)["realm_access"]
	if !ok {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Missing realm access", nil, err)
		c.Abort()
		return
	}
	realmAccess, ok := realmAccessRaw.(map[string]interface{})
	if !ok {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid realm access format", nil, err)
		c.Abort()
		return
	}

	rolesRaw, ok := realmAccess["roles"]
	if !ok {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Missing roles", nil, err)
		c.Abort()
		return
	}
	rolesSlice, ok := rolesRaw.([]interface{})
	if !ok {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid roles format", nil, err)
		c.Abort()
		return
	}
	var matchedRole constant.UserRole
	for _, role := range rolesSlice {
		for _, validRole := range constant.ValidUserRoles {
			if role == string(validRole) {
				matchedRole = validRole
				break
			}
		}
		if matchedRole != "" {
			break
		}
	}

	if matchedRole == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User role not found", nil, err)
		return
	}
	err1 := database.DB.Where("auth_user_id = ?", sub).First(&user).Error
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User not found", nil, err)
		return
	}
	// role, err := uc.roleService.GetRoleByUserId(user.UserId, input.LoginAs)
	// if err != nil {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User role not found", nil, err)
	// 	return
	// }
	logindata := map[string]interface{}{
		"last_login":       time.Now(),
		"first_login_flag": true,
		"user_state":       constant.Active,
		"login_count":      loginInfo.LoginCount + 1,
		"last_login_ip":    c.ClientIP(),
	}
	updateError := uc.userService.UpdateUserInfo(user.AuthUserId, logindata)
	if updateError != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotModified, "User login details not updated", nil, err)
	}
	userLoginResponse := models.UserLoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIN:    token.ExpiresIn,
		UserResponse: models.UserResponse{
			UserId:     user.UserId,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Email:      user.Email,
			Username:   user.Username,
			Role:       string(matchedRole),
			AuthUserId: user.AuthUserId,
		},
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User login successfully", userLoginResponse, nil, nil)
}

func (uc *UserController) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Refresh token is required", nil, err)
		return
	}

	client := config.Client
	ctx := context.Background()

	token, err := client.RefreshToken(ctx, input.RefreshToken, config.KeycloakClientID, config.KeycloakClientSecret, config.KeycloakRealm)
	if err != nil {
		log.Println("Error refreshing token:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to refresh token", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Token refreshed successfully", map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expires_in":    token.ExpiresIn,
	}, nil, nil)
}

func (uc *UserController) LogoutUser(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid refresh token", nil, err)
		return
	}

	client := config.Client
	ctx := context.Background()

	err := client.Logout(ctx, config.KeycloakClientID, config.KeycloakClientSecret, config.KeycloakRealm, input.RefreshToken)
	if err != nil {
		log.Println("Error logging out from Keycloak:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Error while user logout", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User logged out successfully", nil, nil, nil)
}

func (uc *UserController) UserRegisterByPatient(c *gin.Context) {

	sub, patientUserId, isDelegate, err := utils.GetUserIDFromContext(c, uc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.SystemUser_
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if isDelegate {
		reqUserID, err := uc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		if req.RoleName == string(constant.Relative) {
			err = uc.patientService.CanContinue(patientUserId, reqUserID, constant.PermissionAddFamily)
			if err != nil {
				models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "you do not have permission to perform this action", nil, err)
				return
			}
		} else {
			err = uc.patientService.CanContinue(patientUserId, reqUserID, constant.PermissionAddCaregiver)
			if err != nil {
				models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "you do not have permission to perform this action", nil, err)
				return
			}
		}
	} else {
		canAccess := uc.patientService.CanAccessAPI(patientUserId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
		if !canAccess {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Please get subscription inorder to perform this action", nil, errors.New("You need subscription for adding labs"))
			return
		}
	}

	if req.RoleName != string(constant.Relative) && req.RoleName != string(constant.Caregiver) && req.RoleName != string(constant.Doctor) {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid role. Only relative or caregiver can be registered.", nil, nil)
		return
	}
	if err := uc.subscriptionService.ValidateFamilyMemberLimit(patientUserId, req.RoleName); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusOK, err.Error(), nil, err)
		return
	}
	patient, err := uc.patientService.GetUserProfileByUserId(patientUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	registrant := utils.MapSystemUserToPatient(patient)

	relation, err := uc.patientService.GetRelationById(req.RelationId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relation not found", nil, err)
		return
	}

	password := utils.GenerateRandomPassword()
	req.Password = password
	req.Email = strings.ToLower(req.Email)
	req.Username = uc.userService.GenerateUniqueUsername(req.FirstName, req.LastName)
	log.Println("System Generated Password for system user:", password)
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to hash password", nil, err)
		return
	}

	// Get Role ID from role name
	roleMaster, err := uc.roleService.GetRoleIdByRoleName(req.RoleName)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Role not found", nil, err)
		return
	}

	// Create user in Keycloak or authentication system
	keyCloakID, isExistingUser, CreateUserInKeycloakError := uc.authService.CreateUserInKeycloak(req)
	if CreateUserInKeycloakError != nil {
		log.Printf("CreateUserInKeycloak ERROR : %s ", CreateUserInKeycloakError)
	}
	req.Password = string(hashedPassword)
	req.AuthUserId = keyCloakID

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in UserRegisterByPatient:", r)
		}
	}()

	var systemUser models.SystemUser_
	var systemUserErr error
	if !isExistingUser {
		systemUser, systemUserErr = uc.userService.CreateSystemUser(tx, req)
		if systemUserErr != nil {
			tx.Rollback()
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to register user", nil, systemUserErr)
			return
		}
	} else {
		systemUser, systemUserErr = uc.userService.GetSystemUserInfoByAuthUserId(req.AuthUserId)
		if systemUserErr != nil {
			tx.Rollback()
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "User not found with auth user Id", nil, systemUserErr)
			return
		}
	}

	if req.RoleName == string(constant.Relative) {
		err = uc.roleService.AddUserRelativeMappings(tx, patientUserId, systemUser.UserId, relation, roleMaster.RoleId, patient, &systemUser)
		if err != nil {
			tx.Rollback()
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to map user relative to patient", nil, err)
			return
		}
	} else {
		systemUser.RelationId = req.RelationId
		err = uc.roleService.AddSystemUserMapping(tx, &patientUserId, systemUser, patient, roleMaster.RoleId, roleMaster.RoleName, &relation, &isExistingUser, req.RelativeIds)
		if err != nil {
			tx.Rollback()
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to map user caregiver to patient", nil, err)
			return
		}
	}

	if isExistingUser {
		err := uc.emailService.SendConnectionMail(systemUser, registrant, relation.RelationShip)
		if err != nil {
			log.Println("Error sending connection email:", err)
		}
	} else {
		err := uc.emailService.SendLoginCredentials(systemUser, &password, registrant, relation.RelationShip)
		if err != nil {
			log.Println("Error sending email:", err)
		}
	}
	tx.Commit()
	permErr := uc.permissionService.GiveAllPermissionToHOF(&patientUserId, systemUser.UserId)
	if permErr != nil {
		log.Println("GiveAllPermissionToHOF ERROR : ", permErr)
	}
	response := utils.MapUserToRoleSchema(systemUser, roleMaster.RoleName)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User added successfully", response, nil, nil)
	return
}

func (uc *UserController) UserRedirect(c *gin.Context) {
	code := c.Param("code")
	log.Println("Short Code Map:", shortURLMap)
	longURL, ok := shortURLMap[code]
	if !ok {
		c.String(http.StatusNotFound, "Invalid or expired link")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (uc *UserController) CheckUserEmailMobileExist(c *gin.Context) {
	var input models.CheckUserMobileEmail
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}
	result, _, err := uc.userService.CheckUserEmailMobileExist(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to check user contact", nil, err)
		return
	}

	messages := []string{}

	if input.Mobile != "" {
		if result {
			messages = append(messages, "Mobile already exists")
		} else {
			messages = append(messages, "Mobile does not exist")
		}
	}

	if input.Email != "" {
		if result {
			messages = append(messages, "Email already exists")
		} else {
			messages = append(messages, "Email does not exist")
		}
	}

	finalMessage := "No input provided"
	if len(messages) > 0 {
		finalMessage = strings.Join(messages, "; ")
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, finalMessage, result, nil, nil)
}

func ComparePasswords(hashedPassword string, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (ac *UserController) SendResetPasswordLink(c *gin.Context) {

	var req models.CheckUserMobileEmail
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Valid email is required", nil, err)
		return
	}

	err := ac.authService.SendResetPasswordLink(strings.ToLower(req.Email))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to send reset link", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Reset link sent successfully", nil, nil, nil)
}

func (uc *UserController) ResetUserPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Reset token is required", nil, nil)
		return
	}

	tokenData, valid := auth.GetTokenData(token)
	if !valid {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid or expired reset token", nil, nil)
		return
	}
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	userId, err := uc.userService.GetUserIdBySUB(tokenData.AuthUserID)
	if err != nil {
		log.Println("User not found with this authuserId :", tokenData.AuthUserID)
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User not found", nil, nil)
		return
	}
	userInfo, err := uc.patientService.GetUserProfileByUserId(userId)
	if err != nil {
		log.Println("User not found with this ID:", userId)
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User not found", nil, nil)
		return
	}
	if err := auth.ResetPasswordInKeycloak(userInfo.Username, req.Password); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Password reset failed in Keycloak", nil, err)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to hash password", nil, err)
		return
	}
	updateData := map[string]interface{}{
		"password": hashedPassword,
	}
	if err := uc.userService.UpdateUserInfo(userInfo.AuthUserId, updateData); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update password in database", nil, err)
		return
	}
	// auth.DeleteToken(token)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Password reset successfully", nil, nil, nil)
}

func (uc *UserController) FetchAddressByPincode(c *gin.Context) {
	type PostalCodeRequest struct {
		PostalCode string `json:"postal_code" binding:"required"`
	}
	var req PostalCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PostalCode == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Postal code is required", nil, err)
		return
	}

	addressData, err := uc.userService.FetchAddressByPincode(req.PostalCode)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Pincode not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Address fetched successfully", addressData, nil, nil)
}

func (uc *UserController) AddRelationHandler(ctx *gin.Context) {
	sub, patientId, isDelegate, err := utils.GetUserIDFromContext(ctx, uc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.AddRelationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "invalid request body", nil, err)
		return
	}

	if isDelegate {
		reqUserID, err := uc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		if req.RoleName == string(constant.MappingTypeR) {
			err = uc.patientService.CanContinue(patientId, reqUserID, constant.PermissionAddFamily)
			if err != nil {
				models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "you do not have permission to perform this action", nil, err)
				return
			}
		} else {
			err = uc.patientService.CanContinue(patientId, reqUserID, constant.PermissionAddCaregiver)
			if err != nil {
				models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "you do not have permission to perform this action", nil, err)
				return
			}
		}
	} else {
		canAccess := uc.patientService.CanAccessAPI(patientId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
		if !canAccess {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Please get subscription inorder to perform this action", nil, errors.New("You need subscription for updating information"))
			return
		}
	}
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in AddRelationHandler:", r)
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to add relation", nil, nil)
			return
		}
	}()
	mappingErr := uc.patientService.AddRelation(tx, req, patientId)
	if mappingErr != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, mappingErr.Error(), nil, mappingErr)
		return
	}
	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to commit transaction", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Relation added successfully", nil, nil, nil)
	return
}

// func (ac *UserController) SendOTP(c *gin.Context) {
// 	var req models.SendOTPRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Valid 10-digit phone number is required", nil, err)
// 		return
// 	}

// 	err := ac.authService.SendOTP(req.Email)
// 	if err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to send OTP", nil, err)
// 		return
// 	}

// 	models.SuccessResponse(c, constant.Success, http.StatusOK, "OTP sent successfully", nil, nil, nil)
// }

// type OTPVerifyRequest struct {
// 	Email string `json:"email"`
// 	OTP   string `json:"otp"`
// }

// func (ctrl *UserController) VerifyOTP(c *gin.Context) {
// 	var req OTPVerifyRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 		return
// 	}

// 	token, err := ctrl.authService.VerifyOTPAndLogin(req.Email, req.OTP)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "OTP verified", "token": token})
// }

func (uc *UserController) GetAllGender(c *gin.Context) {
	genders, err := uc.patientService.GetAllGender()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch genders", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Genders fetched successfully", genders, nil, nil)
}
