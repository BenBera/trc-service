package auth

import (
	"bitbucket.org/maybets/kra-service/app/constants"
	"bitbucket.org/maybets/kra-service/app/library"
	"bitbucket.org/maybets/kra-service/app/models"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	goutils "github.com/mudphilo/go-utils"
	jwtfiltergolang "github.com/mudphilo/gwt"
	"github.com/sirupsen/logrus"

	"net/http"
	"os"
	"strconv"
	"time"
)

const TokenServiceKey = 1
const TokenTypeAPI = 2
const TokenTypeBasicAuth = 3
const TokenTypeSMPP = 4
const TokenTypeUnknown = 6
const genericAuthFailed = "authorization failed. You are not authorized to %s %s"
const TokenTypeAPIKey = 5

func GetToken(c echo.Context) (token string, tokeType int64) {

	r := c.Request()

	token = r.Header.Get("Authorization")
	if len(token) > 0 {

		return token, TokenTypeAPI
	}

	token = r.Header.Get("x-token")
	if len(token) > 0 {

		return token, TokenServiceKey
	}

	token = r.Header.Get("api-key")
	if len(token) > 0 {

		return token, TokenTypeAPIKey
	}

	return "", TokenTypeUnknown

}

func checkAuthenticate(c echo.Context, module, permission string) (bool, string, int) {

	token, tokenType := GetToken(c)
	var clientID, userID, roleID int64
	roleID = 0

	switch tokenType {

	case TokenServiceKey:

		if token != os.Getenv("SERVICE_TOKEN") {

			return false, "authorization failed, could not retrieve token", http.StatusUnauthorized
		}

		headerClientID, _ := strconv.ParseInt(c.Request().Header.Get("x-client-id"), 10, 64)
		roleID = 1
		clientID = headerClientID
		userID = 1

	case TokenTypeAPI:

		claims, err := jwtfiltergolang.TokenValidation(token)
		if err != nil {
			logrus.WithFields(logrus.Fields{constants.DESCRIPTION: constants.TokenError, constants.DATA: token}).Error(err.Error())
			return false, "authorization failed, could not retieve token", http.StatusUnauthorized
		}

		if module != "self" && permission != "auth" && !jwtfiltergolang.HasPermission(token, module, permission, "ALL") {
			logrus.WithFields(logrus.Fields{constants.DESCRIPTION: fmt.Sprintf("API token %v has not %v permission on %v module ", claims.UserId, permission, module)}).Info()
			return false, fmt.Sprintf(genericAuthFailed, permission, module), http.StatusUnauthorized
		}

		clientID = claims.ClientID
		userID = claims.UserId
		roleID = int64(claims.Role.ID)

	case TokenTypeAPIKey:

		tokenString, err := library.Decrypt(os.Getenv("API_ENCRYPTION_KEY"), token)
		if err != nil {
			logrus.WithFields(logrus.Fields{constants.DESCRIPTION: constants.TokenError, constants.DATA: token}).Error(err.Error())
			return false, "authorization failed, could not retrieve token, token expired", http.StatusUnauthorized
		}

		tokenData := new(models.TokenData)
		err = json.Unmarshal([]byte(tokenString), tokenData)
		if err != nil {

			logrus.WithFields(logrus.Fields{constants.DESCRIPTION: constants.TokenError, constants.DATA: token}).Error(err.Error())
			return false, "authorization failed, could not retrieve token", http.StatusUnauthorized
		}

		if tokenData.Expiry < time.Now().Unix() {

			return false, "Your session has expired, please login again", http.StatusUnauthorized

		}

		// check module permissions
		isAllowed := false

		if module == "self" && permission == "auth" {

			isAllowed = true
		}

		for _, t := range tokenData.Role.Permission {

			if t.Module == module && goutils.Contains(t.Actions, permission) {

				isAllowed = true
			}
		}

		if !isAllowed {

			return false, fmt.Sprintf(genericAuthFailed, permission, module), http.StatusUnauthorized
		}

		clientID = 1
		userID = tokenData.UserID
		roleID = int64(tokenData.Role.ID)

	default:

		return false, constants.AuthorizationFailed, http.StatusUnauthorized

	}

	sess, err := session.Get("session", c)
	if err != nil {

		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: "Session bag error"}).Error(err.Error())
		return false, fmt.Sprintf(genericAuthFailed, permission, module), http.StatusInternalServerError
	}

	sess.Values["client_id"] = clientID
	sess.Values["user_id"] = userID
	sess.Values["role_id"] = roleID
	err = sess.Save(c.Request(), c.Response())
	if err != nil {

		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: "Error saving session error", constants.DATA: token}).Error(err.Error())
		return false, fmt.Sprintf(genericAuthFailed, permission, module), http.StatusInternalServerError
	}

	return true, "", http.StatusOK
}

func Authenticate(pass echo.HandlerFunc, module string, permission string) echo.HandlerFunc {

	return func(c echo.Context) error {

		authenticated, message, httpStatus := checkAuthenticate(c, module, permission)
		if authenticated {

			return pass(c)
		}

		return echo.NewHTTPError(httpStatus, models.ResponseMessage{
			Status:  httpStatus,
			Message: message,
		})
	}
}
