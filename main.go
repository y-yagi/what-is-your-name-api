package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/vision/v1"
	"google.golang.org/appengine"

	"github.com/zenazn/goji"
)

var (
	cfg *jwt.Config
)

func generateFeatures(typeStr string) ([]*vision.Feature, error) {
	types := strings.Split(typeStr, ",")
	features := make([]*vision.Feature, len(types))
	var featureType string

	// Possible values:
	//   "TYPE_UNSPECIFIED" - Unspecified feature type.
	//   "FACE_DETECTION" - Run face detection.
	//   "LANDMARK_DETECTION" - Run landmark detection.
	//   "LOGO_DETECTION" - Run logo detection.
	//   "LABEL_DETECTION" - Run label detection.
	//   "TEXT_DETECTION" - Run OCR.
	//   "SAFE_SEARCH_DETECTION" - Run various computer vision models to
	//   "IMAGE_PROPERTIES" - compute image safe-search properties.
	for i := 0; i < len(types); i++ {
		switch types[i] {
		case "face":
			featureType = "FACE_DETECTION"
		case "landmark":
			featureType = "LANDMARK_DETECTION"
		case "logo":
			featureType = "LOGO_DETECTION"
		case "label":
			featureType = "LABEL_DETECTION"
		case "text":
			featureType = "TEXT_DETECTION"
		case "safe_search":
			featureType = "SAFE_SEARCH_DETECTION"
		case "image_properties":
			featureType = "IMAGE_PROPERTIES"
		default:
			errorMsg := "Invalid feature: " + types[i]
			return nil, errors.New(errorMsg)
		}

		features[i] = &vision.Feature{
			Type:       featureType,
			MaxResults: 5,
		}
	}
	return features, nil
}

func setupConfig() {
	confFile, _ := ioutil.ReadFile("google_credentials.json")
	cfg, _ = google.JWTConfigFromJSON([]byte(confFile), vision.CloudPlatformScope)
}

func hanaInfo(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := cfg.Client(ctx)
	svc, _ := vision.New(client)

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	features, _ := generateFeatures("label")

	enc := base64.StdEncoding.EncodeToString(body)
	img := &vision.Image{Content: enc}
	request := &vision.AnnotateImageRequest{
		Image:    img,
		Features: features,
	}

	batch := &vision.BatchAnnotateImagesRequest{
		Requests: []*vision.AnnotateImageRequest{request},
	}
	res, err := svc.Images.Annotate(batch).Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := json.MarshalIndent(res.Responses[0], "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, string(response))
}

func init() {
	setupConfig()
	http.Handle("/", goji.DefaultMux)
	if os.Getenv("BASIC_AUTH_USER") != "" && os.Getenv("BASIC_AUTH_PASSWORD") != "" {
		goji.Use(BasicAuth)
	}
	goji.Post("/hana/info", hanaInfo)
}
