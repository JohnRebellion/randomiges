package imageRepo

import (
	"context"
	"randomiges/envRouting"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

func Init() error {
	var err error
	cld, err = cloudinary.NewFromParams(envRouting.DevelopmentCloudName, envRouting.DevelopmentApiKey, envRouting.DevelopmentApiSecret)
	return err
}

func UploadImage(file interface{}, id string, tag string) (*uploader.UploadResult, error) {
	return cld.Upload.Upload(context.Background(), file, uploader.UploadParams{PublicID: id, Tags: []string{tag}})
}
