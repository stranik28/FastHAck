package main

import (
	"github.com/ccuetoh/nsfw/nsfw-main/tmp"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

// LinkRequest представляет тело запроса с ссылкой
type LinkRequest struct {
	URL string `json:"url" binding:"required"`
}

// LinkResponse представляет ответ с полученной ссылкой
type LinkResponse struct {
	ReceivedURL string `json:"received_url"`
}

// @title           Gin Swagger Example API
// @version         1.0
// @description     This is a sample server for a simple Gin application with Swagger documentation.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @accept  json
// @produce  json

func main() {
	// Создаем новый роутер Gin
	r := gin.Default()

	// Определяем маршрут для POST-запроса
	r.POST("/link", postLink)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Запускаем сервер на порту 8080
	r.Run(":8007")
}

// postLink обрабатывает POST-запрос для получения ссылки
// @Summary      Получить ссылку
// @Description  Принимает ссылку и возвращает ее в ответе
// @Tags         link
// @Accept       json
// @Produce      json
// @Param        linkRequest  body  LinkRequest  true  "Запрос с ссылкой"
// @Success      200  {object}  LinkResponse  "Полученная ссылка"
// @Failure      400  {object}  gin.H  "Ошибка запроса"
// @Router       /link [post]
func postLink(c *gin.Context) {
	var request LinkRequest

	// Попробуем привязать JSON из тела запроса к структуре
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err, code := tmp.Run(request.URL)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		tmp.ClearDir()
		return
	}
	// Ответим клиенту с полученной ссылкой
	c.JSON(http.StatusOK, gin.H{"Access": code})
	_ = tmp.ClearDir()
}
