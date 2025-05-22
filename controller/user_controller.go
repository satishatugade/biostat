package controller

import (
	"biostat/auth"
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"context"
	"log"
	"strings"

	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	patientService service.PatientService
	roleService    service.RoleService
	userService    service.UserService
	emailService   *service.EmailService
}

// var emailService = service.NewEmailService()

func NewUserController(patientService service.PatientService, roleService service.RoleService,
	userService service.UserService, emailService *service.EmailService) *UserController {
	return &UserController{patientService: patientService, roleService: roleService,
		userService: userService, emailService: emailService}
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

	if user.RoleName == "doctor" || user.RoleName == "caregiver" || user.RoleName == "nurse" || user.RoleName == "pharmacist" {
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

	keyCloakUser := user
	keyCloakUser.Password = rawPassword
	//Add User in Keycloak
	keyCloakID, err := createUserInKeycloak(keyCloakUser)
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
		}
	}()
	systemUser, err := uc.userService.CreateSystemUser(tx, user)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to register user", nil, err)
		return
	}
	log.Println("System User Created in DB with ID:", systemUser.UserId)
	mappingError := uc.roleService.AddSystemUserMapping(tx, nil, systemUser.UserId, nil, roleMaster.RoleId, roleMaster.RoleName, nil)
	if mappingError != nil {
		tx.Rollback()
		log.Println("Error while adding user mapping", mappingError)
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User mapping not added", nil, mappingError)
		return
	}
	tx.Commit()
	err = uc.emailService.SendLoginCredentials(systemUser, rawPassword, nil)
	if err != nil {
		log.Println("Error sending email:", err)
	}
	response := utils.MapUserToRoleSchema(systemUser, roleMaster.RoleName)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User registered successfully", response, nil, nil)
	return
}

func createUserInKeycloak(user models.SystemUser_) (string, error) {
	client := utils.Client
	ctx := context.Background()
	token, err := client.LoginClient(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm)
	if err != nil {
		return "", err
	}
	newuser := gocloak.User{
		Username:      gocloak.StringP(user.Username),
		Email:         gocloak.StringP(user.Email),
		FirstName:     gocloak.StringP(user.FirstName),
		LastName:      gocloak.StringP(user.LastName),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(user.Password),
				Temporary: gocloak.BoolP(false),
			},
		},
		RealmRoles: &[]string{(user.RoleName)},
	}

	role, roleErr := client.GetRealmRole(ctx, token.AccessToken, utils.KeycloakRealm, user.RoleName)
	if roleErr != nil {
		return "User role not found at keycloak server", roleErr
	}

	userID, err := client.CreateUser(ctx, token.AccessToken, utils.KeycloakRealm, newuser)
	if err != nil {
		return "Username or email already exists.", err
	}

	adderr := client.AddRealmRoleToUser(ctx, token.AccessToken, utils.KeycloakRealm, userID, []gocloak.Role{*role})
	if adderr != nil {
		return "Unable to add role to user", adderr
	}
	return userID, nil
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var user models.SystemUser_
	var input struct {
		Username string  `json:"username" binding:"required"`
		Password string  `json:"password" binding:"required"`
		LoginAs  *string `json:"login_as"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Username password required.", nil, err)
		return
	}
	client := utils.Client
	ctx := context.Background()
	token, err := client.Login(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm, input.Username, input.Password)
	if err != nil {
		log.Println("Error logging in to Keycloak:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid user credentials!", nil, err)
		return
	}
	_, claims, err := client.DecodeAccessToken(ctx, token.AccessToken, utils.KeycloakRealm)
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
	err1 := database.DB.Where("auth_user_id = ?", sub).First(&user).Error
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User not found", nil, err)
		return
	}
	log.Println("User login as role ", user.RoleName)
	role, err := uc.roleService.GetRoleByUserId(user.UserId, input.LoginAs)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "User role not found", nil, err)
		return
	}
	userLoginResponse := models.UserLoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		UserResponse: models.UserResponse{
			UserId:     user.UserId,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Email:      user.Email,
			Username:   user.Username,
			Role:       role.RoleName,
			AuthUserId: user.AuthUserId,
		},
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User login successfully", userLoginResponse, nil, nil)
}

func (uc *UserController) LogoutUser(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid refresh token", nil, err)
		return
	}

	client := utils.Client
	ctx := context.Background()

	err := client.Logout(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm, input.RefreshToken)
	if err != nil {
		log.Println("Error logging out from Keycloak:", err)
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Error while user logout", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User logged out successfully", nil, nil, nil)
}

func (uc *UserController) UserRegisterByPatient(c *gin.Context) {

	patientUserId, err := utils.GetUserIDFromContext(c, uc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.SystemUser_
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if req.RoleName != "relative" && req.RoleName != "caregiver" && req.RoleName != "doctor" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid role. Only relative or caregiver can be registered.", nil, nil)
		return
	}

	patient, err := uc.patientService.GetUserProfileByUserId(patientUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	registrant := utils.MapSystemUserToPatient(patient)

	relation, err := uc.patientService.GetRelationById(int(req.RelationId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relation not found", nil, err)
		return
	}

	password := utils.GenerateRandomPassword()
	req.Password = password
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
	keyCloakID, err := createUserInKeycloak(req)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, keyCloakID, nil, err)
		return
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

	systemUser, err := uc.userService.CreateSystemUser(tx, req)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to register user", nil, err)
		return
	}

	relationId := int(relation.RelationId)
	err = uc.roleService.AddSystemUserMapping(tx, &patientUserId, systemUser.UserId, patient, roleMaster.RoleId, roleMaster.RoleName, &relationId)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to map user to patient", nil, err)
		return
	}

	err = uc.emailService.SendLoginCredentials(systemUser, password, registrant)
	if err != nil {
		log.Println("Error sending email:", err)
	}
	log.Println("Email send succesfully ")
	tx.Commit()
	response := utils.MapUserToRoleSchema(systemUser, roleMaster.RoleName)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User registered successfully", response, nil, nil)
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
	result, err := uc.patientService.CheckUserEmailMobileExist(&input)
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

func (uc *UserController) ResetUserPassword(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	if err := auth.ResetPasswordInKeycloak(req.Username, req.NewPassword); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Password reset failed", nil, err)
		return
	}
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
