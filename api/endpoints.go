package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"toto-pizza/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ApiRegister(ctx *gin.Context) {
	var register Register
	var databaseUser database.User
	var users []database.User

	if err := ctx.ShouldBindBodyWithJSON(&register); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	database.DB.Where("phone = ?", register.Phone).Find(&users)
	re := regexp.MustCompile(`^(\+7|7|8)?[\s\-]?\(?[0-9]{3}\)?[\s\-]?[0-9]{3}[\s\-]?[0-9]{2}[\s\-]?[0-9]{2}$`)

	if !re.MatchString(register.Phone) {
		ctx.JSON(401, gin.H{"error": "Incorrect number"})
		return
	}

	if len(users) > 0 {
		ctx.JSON(402, gin.H{"error": "The number is already registered"})
	}

	passwordHash := GetHash(register.Password)

	databaseUser.Balance = 10000
	databaseUser.Name = register.Name
	databaseUser.PasswordHash = passwordHash
	databaseUser.Phone = register.Phone

	database.DB.Create(&databaseUser)
	database.DB.First(&databaseUser, "phone = ?", register.Phone)

	token, err := CreateSession(databaseUser.ID)

	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"token": token})
}

func ApiAuth(ctx *gin.Context) {
	var user database.User
	var authJson Auth

	if err := ctx.ShouldBindBodyWithJSON(&authJson); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	re := regexp.MustCompile(`^(\+7|7|8)?[\s\-]?\(?[0-9]{3}\)?[\s\-]?[0-9]{3}[\s\-]?[0-9]{2}[\s\-]?[0-9]{2}$`)

	if !re.MatchString(authJson.Phone) {
		ctx.JSON(401, gin.H{"error": "Incorrect number"})
		return
	}

	result := database.DB.Where("phone = ?", authJson.Phone).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) && result.Error != nil {
		ctx.JSON(402, gin.H{"error": "User not found"})
		return
	}

	if user.PasswordHash != GetHash(authJson.Password) {
		ctx.JSON(403, gin.H{"error": "Incorrect password"})
		return
	}

	token, _ := CreateSession(user.ID)
	ctx.JSON(http.StatusAccepted, gin.H{"token": token})
}

func ApiGetItems(ctx *gin.Context) {
	var items []database.Item

	result := database.DB.Find(&items)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server error",
		})
		return
	}

	ctx.JSON(http.StatusOK,
		items,
	)
}

func ApiCreateOrder(ctx *gin.Context) {
	var OrderReq CreateOrderReq
	var OrderCost uint
	User := GetUserByToken(ctx.GetHeader("token"))

	if err := ctx.ShouldBindJSON(&OrderReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect body"})
		return
	}

	for _, v := range OrderReq.Items {
		var item database.Item
		database.DB.First(&item, "id = ?", v.Id)
		OrderCost += item.Cost * v.Count
	}

	if User.Balance < OrderCost {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Balnce"})
		return
	}

	var Order database.Order

	Order.Address = OrderReq.Address
	Order.Cost = OrderCost
	Order.Status = "COOK"
	Order.UserId = User.ID

	database.DB.Create(&Order)

	for _, v := range OrderReq.Items {
		var OrderItem database.OrderItem
		var item database.Item
		database.DB.First(&item, "id = ?", v.Id)

		OrderItem.ItemId = v.Id
		OrderItem.OrderId = Order.ID
		OrderItem.Cost = v.Count * item.Cost
		OrderItem.Count = v.Count

		database.DB.Create(&OrderItem)
	}

	User.Balance -= OrderCost
	database.DB.Save(&User)

	ctx.JSON(http.StatusOK, gin.H{"order_id": Order.ID})
}

func ApiGetOrder(ctx *gin.Context) {
	var Order database.Order
	var OrderItems []database.OrderItem
	var Result GetOrderReq
	User := GetUserByToken(ctx.GetHeader("token"))
	OrderId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect number"})
		return
	}

	res := database.DB.First(&Order, "id = ?", OrderId)

	if res.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect ID"})
		return
	}

	if Order.UserId != User.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	database.DB.Where("order_id = ?", Order.ID).Find(&OrderItems)

	Result.Address = Order.Address
	Result.Cost = Order.Cost
	Result.Status = Order.Status

	for _, v := range OrderItems {
		var OrderItem GetOrderItem
		OrderItem.Id = v.ItemId
		OrderItem.Cost = v.Cost
		OrderItem.Count = v.Count

		Result.Items = append(Result.Items, OrderItem)
	}

	ctx.JSON(http.StatusOK, Result)
}

func ApiGetUser(ctx *gin.Context) {
	User := GetUserByToken(ctx.GetHeader("token"))
	ctx.JSON(http.StatusOK, User)
}

func ApiCancelOrder(ctx *gin.Context) {
	User := GetUserByToken(ctx.GetHeader("token"))
	OrderId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	var Order database.Order

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect number"})
		return
	}

	res := database.DB.First(&Order, "id = ?", OrderId)

	if res.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect ID"})
		return
	}

	if Order.UserId != User.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	Order.Status = "CANCEL"

	User.Balance += Order.Cost

	database.DB.Save(&Order)
	database.DB.Save(&User)

	ctx.JSON(http.StatusOK, Order)
}

func ApiGetPhoto(c *gin.Context) {
	id := c.Param("item")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item ID is required"})
		return
	}

	baseDir := "./uploads"
	targetDir := filepath.Join(baseDir)

	// Проверяем существование папки
	fileInfo, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access directory"})
		return
	}

	if !fileInfo.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Uploads path is not a directory"})
		return
	}

	var foundFile string
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(info.Name(), id) {
			foundFile = path
			return io.EOF // Используем io.EOF для досрочного прекращения Walk
		}
		return nil
	})

	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search error: " + err.Error()})
		return
	}

	if foundFile == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Проверяем, что файл существует и доступен
	if _, err := os.Stat(foundFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Found file is not accessible"})
		return
	}

	c.File(foundFile)
}

func ApiGetActiveOrders(ctx *gin.Context) {
	User := GetUserByToken(ctx.GetHeader("token"))
	var Orders []database.Order
	var excludedCategories = []string{"CANCEL", "DONE"}

	result := database.DB.Where("user_id = ?", User.ID).
		Not("status IN ?", excludedCategories).
		Find(&Orders)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if len(Orders) > 0 {
		ctx.JSON(http.StatusOK, Orders)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{"error": "Not found orders"})
	return
}
