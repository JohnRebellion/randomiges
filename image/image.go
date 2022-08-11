package image

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"randomiges/imageRepo"
	"randomiges/user"

	"github.com/JohnRebellion/go-utils/database"
	fiberUtils "github.com/JohnRebellion/go-utils/fiber"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model `json:"-"`
	ID         uint   `json:"id" gorm:"primarykey"`
	Hits       int    `json:"hits"`
	URI        string `json:"uri"`
	Username   string `json:"username"`
}

type ImageData struct {
	ID   uint   `json:"id"`
	Hits int    `json:"hits"`
	URI  string `json:"uri"`
}

type NewImageData struct {
	ID    uint   `json:"id"`
	URI   string `json:"uri"`
	Owner string `json:"owner"`
}

func uploadRandomImages(quantity int, username string) error {
	var err error
	var wg sync.WaitGroup
	j := 0

	for j < quantity {
		wg.Add(1)
		rand.Seed(time.Now().UnixNano())
		maxID := 100000
		minID := 10000
		time.Sleep(time.Millisecond)
		go func() {
			defer wg.Done()
			j++
			uploadResp, err := imageRepo.UploadImage(fmt.Sprintf("https://source.unsplash.com/random/200x200?sig=%d", j), fmt.Sprint(rand.Intn(maxID-minID)+minID), username)

			if err == nil {
				database.DBConn.Create(&Image{
					URI:      uploadResp.SecureURL,
					Username: username,
					Hits:     1,
				})
			}
		}()
	}

	wg.Wait()
	return err
}

type ImageResponse struct {
	Limit int     `json:"limit"`
	Data  []Image `json:"data"`
}

func GetImages(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	limit, err := strconv.Atoi(c.Query("limit", "5"))

	if err != nil {
		return fiberUtils.SendBadRequestResponse("Email has already been registered")
	}

	u, err := getOwner(c)
	err = uploadRandomImages(limit, u.Username)

	if err != nil {
		return fiberUtils.SendBadRequestResponse("Something went wrong")
	}

	images := []Image{}
	err = database.DBConn.Find(&images, "username=?", u.Username).Error

	if err != nil {
		return fiberUtils.SendBadRequestResponse("No images found for user specified")
	} else {
		return c.JSON(&ImageResponse{Limit: limit, Data: images})
	}
}

func GetImage(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	u, err := getOwner(c)
	id, err := getID(c)
	image := new(Image)
	err = database.DBConn.Find(&image, "(username=? OR 'Admin'=?) AND id=?", u.Username, u.Role, id).Error

	if err != nil {
		return fiberUtils.SendBadRequestResponse("No images found for user specified")
	} else {
		image.Hits++
		err = database.DBConn.Updates(&image).Error

		if err != nil {
			return fiberUtils.SendBadRequestResponse("Something went wrong")
		}

		return c.JSON(&ImageData{
			ID:   image.ID,
			Hits: image.Hits,
			URI:  image.URI,
		})
	}
}

func AllImages(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	u, err := getOwner(c)
	images := []Image{}
	err = database.DBConn.Find(&images, "username=? OR 'Admin'=?", u.Username, u.Role).Error

	if err != nil {
		return fiberUtils.SendBadRequestResponse("No images found for user specified")
	} else {
		return c.JSON(images)
	}
}

func DeleteImage(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	u, err := getOwner(c)
	id, err := getID(c)
	err = database.DBConn.Delete(&Image{}, "(username=? OR 'Admin'=?) AND id=?", u.Username, u.Role, id).Error

	if err != nil {
		return fiberUtils.SendBadRequestResponse("No images found for user specified")
	} else {
		return c.SendStatus(http.StatusNoContent)
	}
}

func UpdateImage(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	newImage := new(Image)
	fiberUtils.ParseBody(&newImage)
	u, err := getOwner(c)
	id, err := getID(c)
	image := new(Image)
	err = database.DBConn.Find(&image, "(username=? OR 'Admin'=?) AND id=?", u.Username, u.Role, id).Error

	if err != nil {
		return fiberUtils.SendBadRequestResponse("No images found for user specified")
	}

	if newImage.Username == "" {
		newImage.Username = u.Username
	}

	image.URI = newImage.URI
	image.Hits = newImage.Hits
	err = database.DBConn.Updates(&image).Error

	if err == nil {
		return c.SendStatus(http.StatusAccepted)
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func NewImage(c *fiber.Ctx) error {
	fiberUtils.Ctx.New(c)
	newImageData := new(NewImageData)
	fiberUtils.ParseBody(&newImageData)
	u, err := getOwner(c)
	image := &Image{Hits: 1, URI: newImageData.URI, Username: newImageData.Owner}

	if newImageData.Owner == "" {
		image.Username = u.Username
	}

	err = database.DBConn.Create(&image).Error

	if err == nil {
		return c.SendStatus(http.StatusCreated)
	}

	return fiberUtils.SendJSONMessage("Something went wrong", false, http.StatusInternalServerError)
}

func getOwner(c *fiber.Ctx) (*user.User, error) {
	username := c.Cookies("username")

	if username == "" {
		return nil, errors.New("Invalid cookies")
	}

	u := new(user.User)
	err := database.DBConn.Find(&u, "username=?", username).Error

	if err != nil {
		return nil, errors.New("No users found for user specified")
	}

	return u, err
}

func getID(c *fiber.Ctx) (int, error) {
	id, err := c.ParamsInt("id", 0)

	if err != nil || id == 0 {
		return 0, errors.New("Invalid id")
	}

	return id, err
}
