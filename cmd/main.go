package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	proto "github.com/den-corp/proto"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	conn, err := grpc.Dial("0.0.0.0:2000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := proto.NewNeuroServiceClient(conn)

	g := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"POST", "OPTIONS"}
	g.Use(cors.New(config))

	g.Static("/static", "./static")
	g.LoadHTMLGlob("static/*")

	g.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	g.POST("/upload", func(c *gin.Context) {

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		originalPic, err := c.FormFile("originalPic")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}

		// Open the file
		originalFile, err := originalPic.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}
		defer originalFile.Close()

		// Read the contents of the file into a []byte
		originalPicBytes, err := io.ReadAll(originalFile)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}

		stylePic, err := c.FormFile("stylePic")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}

		// Open the file
		styleFile, err := stylePic.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}
		defer styleFile.Close()

		// Read the contents of the file into a []byte
		stylePicBytes, err := io.ReadAll(styleFile)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}
		response, err := client.GeneratePicture(context.Background(), &proto.GeneratePictureRequest{
			OriginalImage: originalPicBytes,
			StyleImage:    stylePicBytes,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			fmt.Println(err)
			return
		}
		fmt.Println(response.GetResultImage())
		c.Header("Content-Disposition", "attachment; filename=resultPic.png")
		c.Data(http.StatusOK, "application/octet-stream", response.GetResultImage())

	})

	s := &http.Server{
		Addr:         ":2500",
		Handler:      g,
		WriteTimeout: 10 * time.Minute,
	}

	log.Print("starting server on port 2500")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
