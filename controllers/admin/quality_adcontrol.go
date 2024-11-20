// controllers/quality_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/websocket"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET /movies/:movieID/qualities
func GetQualityByMovieEpisodeServer(c *gin.Context) {
	movieID := c.Param("movieID")
	episodeID := c.Param("episodeID")
	serverID := c.Param("serverID")
	qualityCollection := models.GetQualityCollection()

	// Thiết lập ngữ cảnh với thời gian timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Chuyển đổi movieID, episodeID và serverID thành ObjectID
	movieOid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Println("Invalid movie ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided movie ID is not valid",
		})
		return
	}

	episodeOid, err := primitive.ObjectIDFromHex(episodeID)
	if err != nil {
		log.Println("Invalid episode ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided episode ID is not valid",
		})
		return
	}

	serverOid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		log.Println("Invalid server ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided server ID is not valid",
		})
		return
	}
	log.Print("Movie o id:", movieOid)
	// Pipeline để lọc, lookup và sắp xếp
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{
			"movie_id":   movieOid,
			"episode_id": episodeOid,
			"server_id":  serverOid,
			"deleted":    bson.M{"$ne": "deleted"},
		}}},
		bson.D{{"$sort", bson.M{"created_at": 1}}},
	}

	// Thực hiện pipeline
	cursor, err := qualityCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	// Đọc tất cả các quality từ cursor
	var qualities []bson.M
	if err := cursor.All(ctx, &qualities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Lưu kết quả vào Redis cache để tránh truy vấn lại
	cacheKey := "qualities_" + movieID + "_" + episodeID + "_" + serverID
	qualitiesJSON, _ := json.Marshal(qualities)
	dbs.RedisClient.Set(ctx, cacheKey, string(qualitiesJSON), 30*time.Minute)

	// Trả về JSON danh sách qualities
	c.JSON(http.StatusOK, gin.H{"qualities": qualities})
}

// // Thêm quality mới
func AddQuality(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	qualityCollection := models.GetQualityCollection()
	serverCollection := models.GetServerCollection()
	var quality models.Quality

	// Lấy danh sách movie, episode, server từ form
	movieID := c.PostForm("movieid")
	episodeID := c.PostForm("episodeid")
	serverID := c.PostForm("serverid")

	movieoid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid movie ID"})
		return
	}
	quality.MovieID = movieoid

	episodeoid, err := primitive.ObjectIDFromHex(episodeID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid episode ID"})
		return
	}
	quality.EpisodeID = episodeoid

	serveroid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid server ID"})
		return
	}
	quality.ServerID = serveroid

	quality.Title = c.PostForm("title")
	quality.Videourl = c.PostForm("videourl")
	quality.Description = c.PostForm("description")

	// Kiểm tra và chuyển đổi status
	statusStr := c.PostForm("status")
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status"})
		return
	}
	quality.Status = status

	// Validate dữ liệu
	if err := quality.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra tên trùng lặp với cùng movieID
	var existingQuality models.Quality
	filter := bson.M{
		"title":      quality.Title,
		"movie_id":   quality.MovieID,
		"episode_id": quality.EpisodeID,
		"server_id":  quality.ServerID,
		"deleted":    bson.M{"$ne": "deleted"}, // Chỉ lấy các tên chưa bị đánh dấu "deleted"
	}

	if err := qualityCollection.FindOne(ctx, filter).Decode(&existingQuality); err == nil {
		// Nếu tìm thấy tên có cùng số và cùng movieID, trả về lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Quality already exists",
			"message": "An quality with this title already exists for this movie. Please choose a different title.",
		})
		return
	}

	// Gán giá trị ID và trạng thái mặc định
	quality.ID = primitive.NewObjectID()
	quality.CreatedAt = time.Now()
	quality.UpdatedAt = time.Now()

	// Thực hiện thêm quality mới vào MongoDB
	_, err = qualityCollection.InsertOne(ctx, quality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add quality",
			"message": "Unable to add new quality due to a server error. Please try again later.",
		})
		return
	}

	// Tự động thêm ID của tập phim mới vào trường Quality của model Movie
	_, err = serverCollection.UpdateOne(
		ctx,
		bson.M{"_id": quality.ServerID}, // Tìm phim có ServerID khớp với quality.ServerID
		bson.M{"$push": bson.M{"quality": quality.ID}}, // Thêm quality.ID vào trường Quality
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật tập phim vào Server"})
		log.Print(err)
		return
	}

	cacheKey := "qualities_" + movieID + "_" + episodeID + "_" + serverID
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, cacheKey)

	// Gửi thông báo cập nhật qua WebSocket
	message := map[string]interface{}{
		"type":      "quality",
		"message":   "An quality was updated!",
		"movieID":   movieID,
		"episodeID": episodeID,
		"serverID":  serverID,
	}

	// Chuyển thông báo thành JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}

	// Log tin nhắn trước khi gửi
	// log.Println("Broadcasting message:", string(messageJSON))

	// Gửi thông điệp tới tất cả các client qua WebSocket
	websocketServer.BroadcastMessage(messageJSON)
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Quality added successfully!",
	})
}

// DeleteQuality là hàm xử lý yêu cầu xóa ảo
func DeleteQuality(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	qualityCollection := models.GetQualityCollection()
	serverCollection := models.GetServerCollection()
	// Lấy ID từ route
	id := c.Param("id")
	// log.Printf("Received movie ID: %s", id)
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Tìm Quality hiện tại để lấy movieID
	var quality models.Quality
	err = qualityCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&quality)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Quality not found"})
		return
	}
	movieID := quality.MovieID.Hex()     // Lưu movieID dưới dạng chuỗi để gửi qua WebSocket
	episodeID := quality.EpisodeID.Hex() // Lưu episodeID dưới dạng chuỗi để gửi qua WebSocket
	serverID := quality.ServerID.Hex()   // Lưu serverID dưới dạng chuỗi để gửi qua WebSocket
	// Tạo filter để tìm Quality theo ID
	filter := bson.M{"_id": objectID}

	// Tạo update để cập nhật trường Deleted = deleted
	update := bson.M{
		"$set": bson.M{
			"deleted": "deleted",
		},
	}

	// Cập nhật trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = qualityCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Quality can not delete"})
		return
	}

	// Xóa qualityID khỏi trường `quality` của server liên quan
	_, err = serverCollection.UpdateOne(ctx,
		bson.M{"_id": quality.ServerID},
		bson.M{"$pull": bson.M{"quality": objectID}},
	)
	if err != nil {
		log.Printf("Failed to remove quality ID from server %s: %v", quality.ServerID.Hex(), err)
	}

	// Xóa cache trong Redis nếu có sử dụng
	cacheKey := "qualities_" + movieID + "_" + episodeID + "_" + serverID
	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, cacheKey)

	// Gửi thông báo cập nhật qua WebSocket
	message := map[string]interface{}{
		"type":      "quality",
		"message":   "An quality was updated!",
		"movieID":   movieID,
		"episodeID": episodeID,
		"serverID":  serverID,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}
	log.Println("Broadcasting message:", string(messageJSON))
	websocketServer.BroadcastMessage(messageJSON)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Quality delete sucessfully"})
}

// Hàm UpdateQualityField để cập nhật trường cụ thể của Quality
func UpdateQualityField(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Lấy ID của Quality từ tham số URL
	idParam := c.Param("id")
	qualityID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality ID"})
		return
	}

	// Nhận dữ liệu từ request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Kiểm tra xem các trường "field", "value", "movieID", "episodeID", "serverID" có trong request hay không
	field, fieldOk := requestData["field"].(string)
	value, valueOk := requestData["value"]
	movieIDStr, movieIDOk := requestData["movieId"].(string)
	episodeIDStr, episodeIDOk := requestData["episodeId"].(string)
	serverIDStr, serverIDOk := requestData["serverId"].(string)

	if !fieldOk || !valueOk || !movieIDOk || !episodeIDOk || !serverIDOk {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'field', 'value', 'movieID', 'episodeID', or 'serverID'"})
		return
	}

	// Log giá trị 'field', 'value', 'movieID', 'episodeID', 'serverID'
	log.Printf("Field: %s, Value: %+v, MovieID: %s, EpisodeID: %s, ServerID: %s\n", field, value, movieIDStr, episodeIDStr, serverIDStr)

	// Chuyển đổi movieID, episodeID, serverID sang ObjectID
	movieObjectID, err := primitive.ObjectIDFromHex(movieIDStr)
	episodeObjectID, err := primitive.ObjectIDFromHex(episodeIDStr)
	serverObjectID, err := primitive.ObjectIDFromHex(serverIDStr)

	// Kiểm tra trùng lặp title nếu đang cập nhật trường title
	if field == "title" {
		collection := models.GetQualityCollection()

		filter := bson.M{
			"title":      value,
			"movie_id":   movieObjectID,
			"episode_id": episodeObjectID,
			"server_id":  serverObjectID,
			"deleted":    bson.M{"$ne": "deleted"},
			"_id":        bson.M{"$ne": qualityID}, // Loại trừ chính tài liệu đang cập nhật
		}

		var existingQuality models.Quality
		err := collection.FindOne(context.Background(), filter).Decode(&existingQuality)
		if err == nil {
			// Nếu tìm thấy tài liệu trùng lặp
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Duplicate title",
				"message": "A quality with this title already exists for the specified movie, episode, and server.",
			})
			return
		}
	}

	// Chuẩn bị dữ liệu cập nhật cho MongoDB
	updateData := bson.M{
		"$set": bson.M{
			field: value,
		},
	}

	// Cập nhật trường trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := models.GetQualityCollection()
	filter := bson.M{"_id": qualityID}
	_, err = collection.UpdateOne(ctx, filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quality field"})
		return
	}

	// Tạo cacheKey trực tiếp từ chuỗi hex của movieID, episodeID và serverID
	cacheKey := "qualities_" + movieIDStr + "_" + episodeIDStr + "_" + serverIDStr
	dbs.RedisClient.Del(ctx, cacheKey)

	// Gửi thông báo cập nhật qua WebSocket
	message := map[string]interface{}{
		"type":      "quality",
		"message":   "An quality was updated!",
		"movieID":   movieIDStr,
		"episodeID": episodeIDStr,
		"serverID":  serverIDStr,
	}

	// Chuyển thông báo thành JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}

	// Log tin nhắn trước khi gửi
	log.Println("Broadcasting message:", string(messageJSON))

	// Gửi thông điệp tới tất cả các client qua WebSocket
	websocketServer.BroadcastMessage(messageJSON)

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Quality field updated successfully",
	})
}
