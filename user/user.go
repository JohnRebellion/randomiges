package user

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"
	"randomiges/envRouting"
	"randomiges/mailService"
	"time"
	"unicode"

	"github.com/JohnRebellion/go-utils/database"
	fiberUtils "github.com/JohnRebellion/go-utils/fiber"
	"github.com/JohnRebellion/go-utils/passwordHashing"
	"github.com/JohnRebellion/go-utils/twilioService"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model    `json:"-"`
	ID            uint   `json:"id" gorm:"primarykey"`
	Username      string `json:"username" gorm:"unique"`
	Password      string `json:"password"`
	Name          string `json:"name"`
	Role          string `json:"role" gorm:"default:User"`
	ContactNumber string `json:"contactNumber"`
	Email         string `json:"email" gorm:"unique"`
}

func GetUsers(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	users := []User{}
	err := database.DBConn.Find(&users).Error

	if err == nil {
		return c.JSON(users)
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func GetUser(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	user := new(User)
	id, err := c.ParamsInt("id")

	if err == nil {
		err = database.DBConn.Find(&user, id).Error

		if err == nil {
			return c.JSON(user)
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func NewUser(c *fiber.Ctx) error {
	user := new(User)
	fiberUtils.Ctx.New(c)
	err := fiberUtils.ParseBody(user)

	if err == nil {
		err = validateUser(user, false)

		if err == nil {
			return err
		}

		user.Password, err = passwordHashing.HashPassword(user.Password)

		if err == nil {
			err = database.DBConn.Create(&user).Error

			if err == nil {
				return fiberUtils.SendSuccessResponse("User created successfully")
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func validateUser(user *User, ignoreEmail bool) error {
	users := []User{}
	_, err := mail.ParseAddress(user.Username)

	if err != nil {
		return fiberUtils.SendBadRequestResponse("Invalid Email address")
	}

	if err = database.DBConn.Find(&users, "username=?", user.Username).Error; err == nil && !ignoreEmail {
		if len(users) > 0 {
			return fiberUtils.SendBadRequestResponse("Email has already been registered")
		}
	}

	if len(user.Password) < 8 {
		return fiberUtils.SendBadRequestResponse("Must be eight characters long")
	}

	if !contains(user.Password, func(r rune) bool { return unicode.IsUpper(r) && unicode.IsLetter(r) }) {
		return fiberUtils.SendBadRequestResponse("Must contain at least one uppercase character")
	}

	if !contains(user.Password, func(r rune) bool { return unicode.IsLower(r) && unicode.IsLetter(r) }) {
		return fiberUtils.SendBadRequestResponse("Must contain at least one lowercase character")
	}

	if !contains(user.Password, unicode.IsDigit) {
		return fiberUtils.SendBadRequestResponse("Must contain at least one digit")
	}

	if !contains(user.Password, func(r rune) bool { return !unicode.IsDigit(r) && !unicode.IsLetter(r) }) {
		return fiberUtils.SendBadRequestResponse("Must contain at least one special character")
	}

	return err
}

func contains(s string, del func(rune) bool) bool {
	for _, r := range s {
		if del(r) {
			return true
		}
	}

	return false
}

func UpdateUser(c *fiber.Ctx) error {
	user := new(User)
	fiberUtils.Ctx.New(c)
	err := fiberUtils.ParseBody(user)

	if err == nil {
		err = validateUser(user, true)

		if err == nil {
			return err
		}

		user.Password, err = passwordHashing.HashPassword(user.Password)

		if err == nil {
			err = database.DBConn.Updates(&user).Error

			if err == nil {
				return fiberUtils.SendJSONMessage("User created successfully", true, http.StatusAccepted)
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	fiberUtils.Ctx.New(c)

	if err == nil {
		err = database.DBConn.Find(&User{}, id).Error

		if err == nil {
			return fiberUtils.SendSuccessResponse("User deleted successfully")
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func Authenticate(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	user := new(User)
	err := fiberUtils.ParseBody(user)
	password := user.Password

	if err == nil {
		existingUser := new(User)

		if len(user.Username) == 0 || len(password) == 0 {
			return fiberUtils.SendBadRequestResponse("Please Input Username and Password")
		}

		if database.DBConn.Find(&existingUser, "username = ?", user.Username).Error == nil {
			if passwordHashing.CheckPasswordHash(password, existingUser.Password) {
				if otps == nil {
					otps = make(map[string]OTP)
				}

				now := time.Now()
				otps[existingUser.Username] = OTP{User: existingUser, PIN: generatePIN(), Time: &now}
				c.Cookie(&fiber.Cookie{
					Name:  "username",
					Value: existingUser.Username,
				})
				return fiberUtils.SendSuccessResponse("Login successfully")
			} else {
				return fiberUtils.SendJSONMessage("Incorrect Username or Password", false, 401)
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func generatePIN() *string {
	rand.Seed(time.Now().UnixNano())
	pin := fmt.Sprintf("%06d", rand.Intn(100001))
	return &pin
}

func SendOTPBy(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	var err error

	if otp, ok := otps[c.Cookies("username")]; ok {
		pin := *otp.PIN

		if otp.Time.Add(time.Minute*5).Unix() < time.Now().Unix() {
			now := time.Now()
			pin = *generatePIN()
			otps[otp.User.Username] = OTP{User: otp.User, PIN: &pin, Time: &now}
		}

		message := fmt.Sprintf("Your OTP PIN is: %s", pin)

		if c.Params("by") == "sms" {
			if _, err = twilioService.SendSMS(message, otp.User.Username); err == nil {
				return fiberUtils.SendSuccessResponse("Pin successfully sent!")
			}
		}

		if c.Params("by") == "email" {
			_, err := mailService.SendMail(mailService.MailParams{
				From:             mailService.MailOptions.From,
				To:               mailService.MailUser{Email: otp.User.Email, Name: otp.User.Name},
				Subject:          "OTP",
				PlainTextContent: message,
			})

			if err == nil {
				return fiberUtils.SendSuccessResponse("Pin successfully sent!")
			}
		}
	} else {
		return fiberUtils.SendBadRequestResponse("No user logged in!")
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

var otps map[string]OTP

type OTP struct {
	User *User
	PIN  *string
	Time *time.Time
}

type SingleValue struct {
	Value string `json:'value'`
}

func AuthenticatePIN(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	pin := new(SingleValue)
	err := fiberUtils.ParseBody(pin)

	if otps == nil {
		otps = make(map[string]OTP)
	}

	if otp, ok := otps[c.Cookies("username")]; ok {
		if otp.Time.Add(time.Minute*5).Unix() < time.Now().Unix() {
			return fiberUtils.SendBadRequestResponse("Invalid PIN!")
		}

		if *otp.PIN == pin.Value {
			_, err = fiberUtils.GenerateJWTSignedString(fiber.Map{"user": otp.User})

			if err == nil {
				return fiberUtils.SendSuccessResponse("Access Granted!")
			}
		} else {
			return fiberUtils.SendBadRequestResponse("Incorrect PIN!")
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func ResetPassword(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	password := new(SingleValue)
	err := fiberUtils.ParseBody(password)

	if err == nil {
		user := new(User)
		err = database.DBConn.Find(&user, "username = ?", c.Cookies("username")).Error

		if err == nil {
			if resetPasswordRequests == nil {
				resetPasswordRequests = make(map[string]ResetPasswordRequest)
			}

			if resetPasswordRequest, ok := resetPasswordRequests[user.Username]; ok {
				if resetPasswordRequest.Time.Add(time.Hour).Unix() > time.Now().Unix() {
					if c.Cookies("resetToken") == *resetPasswordRequest.Token {
						user.Password, err = passwordHashing.HashPassword(password.Value)

						if err == nil {
							err = database.DBConn.Updates(&user).Error

							if err == nil {
								return fiberUtils.SendSuccessResponse("User created successfully")
							}
						}
					}
				}
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

var resetPasswordRequests map[string]ResetPasswordRequest

type ResetPasswordRequest struct {
	Token *string
	Time  *time.Time
}

func SendResetRequest(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	username := new(SingleValue)
	err := fiberUtils.ParseBody(username)

	if err == nil {
		token, err := passwordHashing.HashPassword(username.Value)

		if err == nil {
			if resetPasswordRequests == nil {
				resetPasswordRequests = make(map[string]ResetPasswordRequest)
			}

			now := time.Now()
			resetPasswordRequests[username.Value] = ResetPasswordRequest{Token: &token, Time: &now}
			user := new(User)
			err = database.DBConn.Find(&user, "username = ?", username.Value).Error

			if err == nil {
				message := fmt.Sprintf("Your Reset Password URL is: %s:%s/api/v1/user/auth/approveRequest?username=%s&token=%s", c.IP(), envRouting.Port, user.Username, token)
				if c.Params("by") == "sms" {
					if _, err = twilioService.SendSMS(message, user.Username); err == nil {
						return fiberUtils.SendSuccessResponse("Reset Password URL successfully sent!")
					}
				}

				if c.Params("by") == "email" {
					if _, err = mailService.SendMail(mailService.MailParams{
						From:             mailService.MailOptions.From,
						To:               mailService.MailUser{Email: user.Email, Name: user.Name},
						Subject:          "Reset Password",
						PlainTextContent: message,
					}); err == nil {
						return fiberUtils.SendSuccessResponse("Reset Password URL successfully sent!")
					}
				}
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func ResetPasswordApprove(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)

	user := new(User)
	err := database.DBConn.Find(&user, "username = ?", c.Query("username")).Error

	if err == nil {
		if resetPasswordRequests == nil {
			resetPasswordRequests = make(map[string]ResetPasswordRequest)
		}
		if resetPasswordRequest, ok := resetPasswordRequests[user.Username]; ok {
			if resetPasswordRequest.Time.Add(time.Hour).Unix() > time.Now().Unix() {
				if c.Query("token") == *resetPasswordRequest.Token {
					c.Cookie(&fiber.Cookie{
						Name:  "resetToken",
						Value: c.Query("token"),
					})
					c.Cookie(&fiber.Cookie{
						Name:  "username",
						Value: user.Username,
					})
					return c.Redirect("/#resetPassword")
				} else {
					return fiberUtils.SendBadRequestResponse("Password reset invalid")
				}
			}
		}
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}
