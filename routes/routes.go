package routes

import (
	"bytes"
	"fmt"
	"illusionman1212/kotapi-go/db"
	"illusionman1212/kotapi-go/models"
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edwvee/exiffix"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomHandler(w http.ResponseWriter, req *http.Request) {
	var kot models.Kot

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	max, err := db.Kots.EstimatedDocumentCount(db.Ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An error has occurred while querying the database", "status": 500, "failed": true }`))
		return
	}

	// ids start from 1
	min := 1
	// seed the default source using the current time in unix nanosecond format
	rand.Seed(time.Now().UnixNano())
	// generate random number between min and `max`
	num := rand.Intn(int(max)-min) + min

	// find the random kot document and store it in the `kot` variable
	err = db.Kots.FindOne(db.Ctx, bson.M{"id": num}).Decode(&kot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "error": "Not found","status": 404, "failed": true }`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "error": "An unknown error has occurred","status": 500, "failed": true }`))
		}
		return
	}

	// create a json from the kot struct variable and then write it as a response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"id": "` + strconv.Itoa(int(kot.MyID)) + `",
		"url": "` + kot.Url + `",
		"compressed_url": "` + kot.CompressedUrl + `",
		"status": 200,
		"failed": false
	}`))
}

func IdHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var kot models.Kot

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "error": "Invalid or incomplete request","status": 400, "failed": true }`))
		return
	}

	err = db.Kots.FindOne(db.Ctx, bson.M{"id": id}).Decode(&kot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "error": "Not found","status": 404, "failed": true }`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "error": "An unknown error has occurred","status": 500, "failed": true }`))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"id": "` + strconv.Itoa(int(kot.MyID)) + `",
		"url": "` + kot.Url + `",
		"compressed_url": "` + kot.CompressedUrl + `",
		"status": 200,
		"failed": false
	}`))
}

func AddKotHandler(w http.ResponseWriter, req *http.Request) {
	var kot models.Kot

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	// 10 MB file size limit
	req.ParseMultipartForm(10 << 20)

	file, handler, err := req.FormFile("image")
	if err != nil {
		log.Print("no image was uploaded")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "error": "Invalid or incomplete request","status": 400, "failed": true }`))
		return
	}

	password := req.Form.Get("password")
	if password == "" {
		log.Print("no password was sent")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "error": "Invalid or incomplete request","status": 400, "failed": true }`))
		return
	}
	if password != os.Getenv("PASSWORD") {
		log.Print("password is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{ "error": "Not authorized","status": 401, "failed": true }`))
		return
	}

	filename := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	_, err = rand.Read(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An unknown error has occurred","status": 500, "failed": true }`))
		return
	}

	extensions := strings.Split(handler.Filename, ".")
	extension := strings.ToLower(extensions[len(extensions)-1])

	defer file.Close()

	data := make([]byte, handler.Size)
	file.Read(data)
	compressedData := data

	if extension == "png" {
		enc := &png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		buf := &bytes.Buffer{}
		decodedImage, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "error": "An error occurred while compressing the image","status": 500, "failed": true }`))
			return
		}
		enc.Encode(buf, decodedImage)
		compressedData = buf.Bytes()
	} else if extension == "jpeg" || extension == "jpg" {
		decodedImage, _, err := exiffix.Decode(bytes.NewReader(data))
		buf := &bytes.Buffer{}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "error": "An error occurred while compressed the image","status": 500, "failed": true }`))
			return
		}

		jpeg.Encode(buf, decodedImage, &jpeg.Options{Quality: 75})
		compressedData = buf.Bytes()
	}

	os.Mkdir("./kots", 0755)
	os.Mkdir("./kots/compressed", 0755)
	newFile, err := os.Create(fmt.Sprintf("./kots/%x.%s", filename, extension))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An error occurred while writing the file","status": 500, "failed": true }`))
		return
	}

	defer newFile.Close()
	newFile.Write(data)

	compressedNewFile, err := os.Create(fmt.Sprintf("./kots/compressed/%x.%s", filename, extension))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An error occurred while writing compressed file","status": 500, "failed": true }`))
		return
	}

	defer compressedNewFile.Close()
	compressedNewFile.Write(compressedData)

	count, err := db.Kots.EstimatedDocumentCount(db.Ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An unknown error has occurred","status": 500, "failed": true }`))
		return
	}

	kot.ID = primitive.NewObjectIDFromTimestamp(time.Now())
	kot.MyID = int32(count + 1)
	kot.Url = fmt.Sprintf("%s/kots/%x.%s", os.Getenv("BASE_URL"), filename, extension)
	kot.CompressedUrl = fmt.Sprintf("%s/kots/compressed/%x.%s", os.Getenv("BASE_URL"), filename, extension)

	_, err = db.Kots.InsertOne(db.Ctx, kot)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "error": "An unknown error has occurred","status": 500, "failed": true }`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "message": "Successfully welcomed a new kot :)" "status": 200, "failed": false }`))
}

func KotsHandler(w http.ResponseWriter, req *http.Request) {
	filename := mux.Vars(req)
	filepath := fmt.Sprintf("./kots/%s", filename["filename"])
	_, err := os.Stat(filepath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{ "error": "Not found","status": 404, "failed": true }`))
		return
	}
	http.ServeFile(w, req, filepath)
}

func KotsCompressedHandler(w http.ResponseWriter, req *http.Request) {
	filename := mux.Vars(req)
	filepath := fmt.Sprintf("./kots/compressed/%s", filename["filename"])
	_, err := os.Stat(filepath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{ "error": "Not found","status": 404, "failed": true }`))
		return
	}
	http.ServeFile(w, req, filepath)
}
