// controllers/episode_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fire-watch/websocket"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET /movies/:movieID/episodes
func GetEpisodesByMovieID(c *gin.Context) {
	movieID := c.Param("movieID")
	episodeCollection := models.GetEpisodeCollection()

	// Thiết lập ngữ cảnh với thời gian timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Chuyển đổi movieID thành ObjectID
	oid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Println("Invalid movie ID:", err) // Log lỗi nếu ObjectID không hợp lệ
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "The provided movie ID is not valid",
		})
		return
	}
	// log.Println("Converted movie ID to ObjectID:", oid) // Log ObjectID đã chuyển đổi

	// Pipeline để lọc, lookup và sắp xếp
	pipeline := mongo.Pipeline{
		// Điều kiện lọc movieid và trạng thái xóa (Deleted != "deleted")
		bson.D{{"$match", bson.M{
			"movieid": oid,
			"deleted": bson.M{"$ne": "deleted"},
		}}},

		// Lookup để nối với collection servers
		bson.D{{"$lookup", bson.M{
			"from":         "servers",
			"localField":   "server",
			"foreignField": "_id",
			"as":           "server_details",
		}}},
		bson.D{{"$addFields", bson.D{
			{"server_details", bson.D{
				{"$filter", bson.D{
					{"input", "$server_details"},
					{"as", "server"},
					{"cond", bson.D{{"$ne", bson.A{"$$server.deleted", "deleted"}}}},
				}},
			}},
		}}},

		// Sắp xếp theo thời gian tạo mới nhất
		bson.D{{"$sort", bson.M{"created_at": 1}}},
	}

	// Thực hiện pipeline
	cursor, err := episodeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	// Đọc tất cả các episode từ cursor
	var episodes []bson.M
	if err := cursor.All(ctx, &episodes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Lưu kết quả vào Redis cache để tránh truy vấn lại
	episodesJSON, _ := json.Marshal(episodes)
	dbs.RedisClient.Set(ctx, "episodes_"+movieID, string(episodesJSON), 30*time.Minute)

	// Trả về JSON danh sách episodes
	c.JSON(http.StatusOK, gin.H{"episodes": episodes})
}

// Thêm episode mới
func AddEpisode(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo biến chứa dữ liệu từ form
	movieCollection := models.GetMovieCollection()
	episodeCollection := models.GetEpisodeCollection()
	serverCollection := models.GetServerCollection()
	var episode models.Episode

	// Lấy danh sách movie từ form
	movieID := c.PostForm("movieid")
	log.Printf("Received movie ID: %s", movieID)
	movieoid, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid movie ID"})
		return
	}
	episode.MovieID = movieoid

	number, _ := strconv.Atoi(c.PostForm("number"))
	episode.Number = number
	status, _ := strconv.Atoi(c.PostForm("status"))
	episode.Status = status

	// Lấy danh sách server từ form
	serverIDs := c.PostFormArray("server[]")
	for _, id := range serverIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid server ID"})
			return
		}
		episode.Server = append(episode.Server, oid)
	}

	// Validate dữ liệu episode
	if err := episode.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(), // Trả về danh sách các lỗi chi tiết
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kiểm tra tập trùng lặp với cùng movieID
	var existingEpisode models.Episode
	filter := bson.M{
		"number":  episode.Number,
		"movieid": episode.MovieID,          // Kiểm tra cả số tập và movieID
		"deleted": bson.M{"$ne": "deleted"}, // Chỉ lấy các tập chưa bị đánh dấu "deleted"
	}

	if err := episodeCollection.FindOne(ctx, filter).Decode(&existingEpisode); err == nil {
		// Nếu tìm thấy tập có cùng số và cùng movieID, trả về lỗi
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Episode already exists",
			"message": "An episode with this number already exists for this movie. Please choose a different number.",
		})
		return
	}

	// Lấy file upload từ form
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Image file is required"})
		return
	}

	// Kiểm tra định dạng ảnh
	allowedFormats := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
	if !allowedFormats[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid image format. Only JPEG WEBP and PNG are allowed."})
		return
	}

	// Kiểm tra kích thước ảnh (giới hạn 2MB)
	const maxImageSize = 2 * 1024 * 1024 // 2MB
	if file.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Image file is too large. Maximum size is 2MB."})
		return
	}

	imageFileName := file.Filename
	imagePath := fmt.Sprintf("views/uploads/images/%s", imageFileName)

	// Kiểm tra nếu file đã tồn tại
	if _, err := os.Stat(imagePath); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Image file already exists. Please upload a different file.",
		})
		return
	} else if !os.IsNotExist(err) {
		// Nếu lỗi khác không phải là "không tồn tại"
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to check existing file"})
		return
	}

	// Kiểm tra thành công, tiến hành mở file
	srcFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to open image"})
		return
	}
	defer srcFile.Close()

	// Gán giá trị ID và trạng thái mặc định
	episode.ID = primitive.NewObjectID()
	episode.CreatedAt = time.Now()
	episode.UpdatedAt = time.Now()

	// Lưu tên file (không lưu đường dẫn đầy đủ)
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		log.Println("Failed to save image file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save image"})
		return
	}
	episode.Image = imageFileName

	// Thực hiện thêm episode mới vào MongoDB
	_, err = episodeCollection.InsertOne(ctx, episode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add episode",
			"message": "Unable to add new episode due to a server error. Please try again later.",
		})
		return
	}

	// Tự động thêm ID của tập phim mới vào trường Episode của model Movie
	_, err = movieCollection.UpdateOne(
		ctx,
		bson.M{"_id": episode.MovieID}, // Tìm phim có MovieID khớp với episode.MovieID
		bson.M{"$push": bson.M{"episode": episode.ID}}, // Thêm episode.ID vào trường Episode
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật tập phim vào Movie"})
		log.Print(err)
		return
	}

	// Cập nhật các Server liên quan để thêm MovieID và EpisodeID vào các trường tương ứng
	_, err = serverCollection.UpdateMany(
		ctx,
		bson.M{"_id": bson.M{"$in": episode.Server}}, // Lọc server bằng danh sách ID trong episode.Server
		bson.M{
			"$addToSet": bson.M{
				"movie_ids":   episode.MovieID, // Thêm movie ID vào trường MovieIDs
				"episode_ids": episode.ID,      // Thêm episode ID vào trường EpisodeIDs
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update servers with new Episode ID"})
		log.Print(err)
		return
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "episodes")
	dbs.RedisClient.Del(ctx, "movies")

	// Gửi thông báo cập nhật episode với movieID
	message := map[string]interface{}{
		"type":    "episode",
		"message": "A new episode was updated!",
		"movieID": movieID,
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
	// Trả về thông báo thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "Episode added successfully!",
	})
}

func UpdateEpisode(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	// Khởi tạo kết nối tới các collection
	episodeCollection := models.GetEpisodeCollection()
	serverCollection := models.GetServerCollection()

	// Nhận dữ liệu từ form
	episodeID := c.PostForm("id")
	movieID := c.PostForm("movieid")
	log.Printf("Received movie ID: %s, episode ID: %s", movieID, episodeID)

	// Chuyển đổi ID sang ObjectID
	episodeOID, err := primitive.ObjectIDFromHex(episodeID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid episode ID"})
		return
	}
	movieOID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid movie ID"})
		return
	}

	// Tìm episode hiện tại để lấy thông tin ảnh cũ
	var existingEpisode models.Episode
	if err := episodeCollection.FindOne(context.TODO(), bson.M{"_id": episodeOID}).Decode(&existingEpisode); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Episode not found"})
		return
	}

	// Nhận các thông tin khác từ form
	number, _ := strconv.Atoi(c.PostForm("number"))
	status, _ := strconv.Atoi(c.PostForm("status"))

	// Lấy file ảnh mới từ form
	file, err := c.FormFile("image")
	var newImageFileName string
	if err == nil { // Nếu có file mới
		allowedFormats := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
		if !allowedFormats[file.Header.Get("Content-Type")] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid image format. Only JPEG, WEBP, and PNG are allowed."})
			return
		}
		if file.Size > 2*1024*1024 { // Kiểm tra kích thước ảnh
			c.JSON(http.StatusBadRequest, gin.H{"message": "Image file is too large. Maximum size is 2MB."})
			return
		}

		newImageFileName = file.Filename
		newImagePath := fmt.Sprintf("views/uploads/images/%s", newImageFileName)

		// Kiểm tra nếu file đã tồn tại
		if _, err := os.Stat(newImagePath); err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "Image file already exists. Please upload a different file."})
			return
		}

		// Lưu file mới
		if err := c.SaveUploadedFile(file, newImagePath); err != nil {
			log.Println("Failed to save new image file:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save new image"})
			return
		}

		// Xóa ảnh cũ nếu có
		oldImagePath := fmt.Sprintf("views/uploads/images/%s", existingEpisode.Image)
		if existingEpisode.Image != "" && existingEpisode.Image != newImageFileName {
			if err := os.Remove(oldImagePath); err != nil {
				log.Printf("Failed to delete old image: %s", oldImagePath)
			}
		}
	}

	// Lấy danh sách server từ form
	serverIDs := c.PostFormArray("server[]")
	var servers []primitive.ObjectID
	for _, id := range serverIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid server ID"})
			return
		}
		servers = append(servers, oid)
	}

	// Xác định các server cần xóa EpisodeID khỏi trường `EpisodeIDs`
	serversToRemove := difference(existingEpisode.Server, servers)
	log.Printf("Servers to remove EpisodeID %s: %v", episodeOID.Hex(), serversToRemove)

	// Xóa EpisodeID khỏi các servers không còn được liên kết
	for _, serverID := range serversToRemove {
		_, err := serverCollection.UpdateOne(context.TODO(),
			bson.M{"_id": serverID},
			bson.M{"$pull": bson.M{"episode_ids": episodeOID}}, // Xóa episodeOID khỏi mảng `EpisodeIDs`
		)
		if err != nil {
			log.Printf("Failed to remove EpisodeID from server %s: %v", serverID.Hex(), err)
		}
	}

	// Kiểm tra trùng lặp tập với cùng movieID
	filter := bson.M{
		"number":  number,
		"movieid": movieOID,
		"_id":     bson.M{"$ne": episodeOID},
	}
	if err := episodeCollection.FindOne(context.TODO(), filter).Decode(&existingEpisode); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Episode already exists",
			"message": "An episode with this number already exists for this movie. Please choose a different number.",
		})
		return
	}

	// Chuẩn bị dữ liệu để cập nhật episode
	update := bson.M{
		"number":    number,
		"status":    status,
		"server":    servers,
		"updatedat": time.Now(),
	}
	if newImageFileName != "" {
		update["image"] = newImageFileName
	}

	// Thực hiện cập nhật episode
	_, err = episodeCollection.UpdateOne(context.TODO(), bson.M{"_id": episodeOID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update episode"})
		return
	}

	// Cập nhật EpisodeID vào các server mới nếu chưa tồn tại
	for _, serverID := range servers {
		_, err := serverCollection.UpdateOne(context.TODO(),
			bson.M{"_id": serverID},
			bson.M{"$addToSet": bson.M{"episode_ids": episodeOID}}, // Thêm EpisodeID nếu chưa có
		)
		if err != nil {
			log.Printf("Failed to add EpisodeID to server %s: %v", serverID.Hex(), err)
		}
	}

	// Xóa cache trong Redis
	dbs.RedisClient.Del(context.TODO(), "episodes")

	// Gửi thông báo cập nhật qua WebSocket
	message := map[string]interface{}{
		"type":    "episode",
		"message": "An episode was updated!",
		"movieID": movieID,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}
	log.Println("Broadcasting message:", string(messageJSON))
	websocketServer.BroadcastMessage(messageJSON)

	// Trả về phản hồi thành công
	c.JSON(http.StatusOK, gin.H{"message": "Episode updated successfully!"})
}

// DeleteEpisode là hàm xử lý yêu cầu xóa ảo
func DeleteEpisode(c *gin.Context, websocketServer *websocket.WebSocketServer) {
	episodeCollection := models.GetEpisodeCollection()
	movieCollection := models.GetMovieCollection()
	serverCollection := models.GetServerCollection()
	// Lấy ID từ route
	id := c.Param("id")
	// log.Printf("Received movie ID: %s", id)
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Tìm Episode hiện tại để lấy movieID
	var episode models.Episode
	err = episodeCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&episode)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Episode not found"})
		return
	}
	movieID := episode.MovieID.Hex() // Lưu movieID dưới dạng chuỗi để gửi qua WebSocket
	movieIDs := episode.MovieID      // Lưu movieID để dùng trong update movie
	serverIDs := episode.Server      // Lưu danh sách server để xóa EpisodeID trong Server
	// Tạo filter để tìm Episode theo ID
	filter := bson.M{"_id": objectID}

	// Tạo update để cập nhật trường Deleted = 1
	update := bson.M{
		"$set": bson.M{
			"deleted": "deleted",
		},
	}

	// Xóa ảnh chính nếu có
	if episode.Image != "" {
		mainImagePath := fmt.Sprintf("views/uploads/images/%s", episode.Image)
		if err := os.Remove(mainImagePath); err != nil && !os.IsNotExist(err) {
			log.Println("Failed to delete main image:", err)
		}
	}

	// Cập nhật trong MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = episodeCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Episode can not delete"})
		return
	}

	// Xóa episodeID khỏi trường `Episodes` của bảng Movie
	_, err = movieCollection.UpdateOne(ctx,
		bson.M{"_id": movieIDs},
		bson.M{"$pull": bson.M{"episode": objectID}},
	)
	if err != nil {
		log.Printf("Failed to remove episode ID from movie %s: %v", movieIDs.Hex(), err)
	}

	// Xóa episodeID khỏi trường `EpisodeIDs` của các server liên quan
	for _, serverID := range serverIDs {
		_, err := serverCollection.UpdateOne(ctx,
			bson.M{"_id": serverID},
			bson.M{"$pull": bson.M{"episode_ids": objectID}},
		)
		if err != nil {
			log.Printf("Failed to remove episode ID from server %s: %v", serverID.Hex(), err)
		}
	}

	// Xóa cache trong Redis nếu có sử dụng
	dbs.RedisClient.Del(ctx, "episodes")

	// Gửi thông báo cập nhật qua WebSocket
	message := map[string]interface{}{
		"type":    "episode",
		"message": "An episode was updated!",
		"movieID": movieID,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error encoding JSON message:", err)
		return
	}
	log.Println("Broadcasting message:", string(messageJSON))
	websocketServer.BroadcastMessage(messageJSON)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Episode delete sucessfully"})
}

// Hàm phụ trợ xác định các server cần loại bỏ EpisodeID
func difference(oldServers, servers []primitive.ObjectID) []primitive.ObjectID {
	var diff []primitive.ObjectID
	oldMap := make(map[primitive.ObjectID]bool)
	for _, id := range oldServers {
		oldMap[id] = true
	}
	for _, id := range servers {
		delete(oldMap, id)
	}
	for id := range oldMap {
		diff = append(diff, id)
	}
	return diff
}
