// controllers/country_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Thêm country mới
func AddCountry(c *gin.Context) {
	countryCollection := models.GetCountryCollection()
	
	var country models.Country
	if err := c.ShouldBindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country data"})
		return
	}

	country.ID = primitive.NewObjectID()
	country.Status = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := countryCollection.InsertOne(ctx, country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting country"})
		return
	}

	// Xóa cache Redis sau khi thêm country mới
	dbs.RedisClient.Del(ctx, "countries")

	c.JSON(http.StatusOK, country)
}

// Lấy tất cả countries
func GetAllCountries(c *gin.Context) {
	countryCollection := models.GetCountryCollection()
	var countries []models.Country

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra dữ liệu cache từ Redis
	cachedCountries, err := dbs.RedisClient.Get(ctx, "countries").Result()
	if err == nil && cachedCountries != "" {
		// Nếu có cache, giải mã từ Redis cache
		json.Unmarshal([]byte(cachedCountries), &countries)
		c.JSON(http.StatusOK, countries)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	cursor, err := countryCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching countries"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var country models.Country
		if err := cursor.Decode(&country); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding country"})
			return
		}
		countries = append(countries, country)
	}

	// Lưu dữ liệu vào Redis cache
	countriesJSON, _ := json.Marshal(countries)
	dbs.RedisClient.Set(ctx, "countries", string(countriesJSON), 30*time.Minute)

	c.JSON(http.StatusOK, countries)
}

// Lấy một country theo ID
func GetCountryByID(c *gin.Context) {
	countryCollection := models.GetCountryCollection()
	id := c.Query("id")
	countryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra cache từ Redis
	cachedCountry, err := dbs.RedisClient.Get(ctx, "country_"+id).Result()
	if err == nil && cachedCountry != "" {
		// Nếu có cache, trả về cache
		var country models.Country
		json.Unmarshal([]byte(cachedCountry), &country)
		c.JSON(http.StatusOK, country)
		return
	}

	// Nếu không có cache, lấy từ MongoDB
	var country models.Country
	err = countryCollection.FindOne(ctx, bson.M{"_id": countryID}).Decode(&country)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Country not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching country"})
		}
		return
	}

	// Lưu dữ liệu vào Redis cache
	countryJSON, _ := json.Marshal(country)
	dbs.RedisClient.Set(ctx, "country_"+id, string(countryJSON), 30*time.Minute)

	c.JSON(http.StatusOK, country)
}

// Cập nhật country
func UpdateCountry(c *gin.Context) {
	countryCollection := models.GetCountryCollection()
	var country models.Country
	if err := c.ShouldBindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country data"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": country.ID}
	update := bson.M{"$set": country}

	_, err := countryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating country"})
		return
	}

	// Xóa cache Redis sau khi cập nhật
	dbs.RedisClient.Del(ctx, "country_"+country.ID.Hex())
	dbs.RedisClient.Del(ctx, "countries")

	c.JSON(http.StatusOK, gin.H{"message": "Country updated successfully"})
}

// Xóa country
func DeleteCountry(c *gin.Context) {
	movieCollection := models.GetMovieCollection()
	countryCollection := models.GetCountryCollection()
	id := c.Param("id")
	countryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Xóa country trong countryCollection
	_, err = countryCollection.DeleteOne(ctx, bson.M{"_id": countryID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting country"})
		return
	}

	// Cập nhật tất cả các bộ phim, xóa countryID khỏi trường Country của Movie
	_, err = movieCollection.UpdateMany(
		ctx,
		bson.M{"country": countryID},            // Tìm những bộ phim chứa countryID
		bson.M{"$unset": bson.M{"country": ""}}, // Xóa giá trị country (đặt về null)
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movies"})
		return
	}
	// Xóa cache Redis sau khi xóa
	dbs.RedisClient.Del(ctx, "countries_"+id)
	dbs.RedisClient.Del(ctx, "countries")
	dbs.RedisClient.Del(ctx, "movies_cache")

	c.JSON(http.StatusOK, gin.H{"message": "country deleted and movies updated successfully"})
}
