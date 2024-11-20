// controllers/movie_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fire-watch/dbs"
	"fire-watch/models"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMoviesDetail(c *gin.Context) (*models.Movie, error) {
	// Lấy collection Movie từ MongoDB
	movieCollection := models.GetMovieCollection()

	// Lấy ID từ route parameter
	id := c.Param("id")
	movieID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("Invalid movie ID: %v", err)
	}

	// Tạo context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Khóa cache Redis
	cacheKey := "movie_detail_" + id

	// Kiểm tra dữ liệu trong Redis cache
	cachedMovie, err := dbs.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Cache không tồn tại, lấy dữ liệu từ MongoDB qua Aggregation Pipeline
		pipeline := mongo.Pipeline{
			// Match movie by ID and exclude deleted
			bson.D{{"$match", bson.D{{"_id", movieID}, {"deleted", bson.D{{"$ne", "deleted"}}}}}},

			// Lookup genres
			bson.D{{"$lookup", bson.D{
				{"from", "genres"},
				{"localField", "genre"},
				{"foreignField", "_id"},
				{"as", "genreDetails"},
			}}},

			// Filter genres to exclude deleted
			bson.D{{"$addFields", bson.D{
				{"genreDetails", bson.D{
					{"$filter", bson.D{
						{"input", bson.D{
							{"$sortArray", bson.D{
								{"input", "$genreDetails"},
								{"sortBy", bson.D{{"CreatedAt", -1}}},
							}},
						}},
						{"as", "genre"},
						{"cond", bson.D{
							{"$and", bson.A{
								bson.D{{"$ne", bson.A{"$$genre.deleted", "deleted"}}},
								bson.D{{"$ne", bson.A{"$$genre.status", 2}}},
							}},
						}},
					}},
				}},
			}}},

			// Lookup episodes related to the movie
			bson.D{{"$lookup", bson.D{
				{"from", "episodes"},
				{"localField", "episode"},
				{"foreignField", "_id"},
				{"as", "episodeDetails"},
			}}},

			// Filter episodes to exclude deleted
			bson.D{{"$addFields", bson.D{
				{"episodeDetails", bson.D{
					{"$filter", bson.D{
						{"input", bson.D{
							{"$sortArray", bson.D{
								{"input", "$episodeDetails"},
								{"sortBy", bson.D{{"CreatedAt", -1}}},
							}},
						}},
						{"as", "episode"},
						{"cond", bson.D{
							{"$and", bson.A{
								bson.D{{"$ne", bson.A{"$$episode.deleted", "deleted"}}},
								bson.D{{"$ne", bson.A{"$$episode.status", 2}}},
							}},
						}},
					}},
				}},
			}}},

			// Lookup servers for all episodes
			bson.D{{"$lookup", bson.D{
				{"from", "servers"},
				{"localField", "episodeDetails.server"},
				{"foreignField", "_id"},
				{"as", "serverDetails"},
			}}},

			// Add server details to episodes
			bson.D{{"$addFields", bson.D{
				{"episodeDetails", bson.D{
					{"$map", bson.D{
						{"input", "$episodeDetails"},
						{"as", "episode"},
						{"in", bson.D{
							{"$mergeObjects", bson.A{
								"$$episode",
								bson.D{{"serverDetails", bson.D{
									{"$filter", bson.D{
										{"input", bson.D{
											{"$sortArray", bson.D{
												{"input", "$serverDetails"},
												{"sortBy", bson.D{{"CreatedAt", -1}}},
											}},
										}},
										{"as", "server"},
										{"cond", bson.D{
											{"$and", bson.A{
												bson.D{{"$in", bson.A{"$$server._id", "$$episode.server"}}},
												bson.D{{"$ne", bson.A{"$$server.deleted", "deleted"}}},
												bson.D{{"$ne", bson.A{"$$server.status", 2}}},
											}},
										}},
									}},
								}}},
							}},
						}},
					}},
				}},
			}}},

			// Lookup qualities for all servers
			bson.D{{"$lookup", bson.D{
				{"from", "qualities"},
				{"localField", "episodeDetails.serverDetails.quality"},
				{"foreignField", "_id"},
				{"as", "qualityDetails"},
			}}},

			// Add quality details to servers
			bson.D{{"$addFields", bson.D{
				{"episodeDetails", bson.D{
					{"$map", bson.D{
						{"input", "$episodeDetails"},
						{"as", "episode"},
						{"in", bson.D{
							{"$mergeObjects", bson.A{
								"$$episode",
								bson.D{{"serverDetails", bson.D{
									{"$map", bson.D{
										{"input", "$$episode.serverDetails"},
										{"as", "server"},
										{"in", bson.D{
											{"$mergeObjects", bson.A{
												"$$server",
												bson.D{{"qualityDetails", bson.D{
													{"$filter", bson.D{
														{"input", bson.D{
															{"$sortArray", bson.D{
																{"input", "$qualityDetails"},
																{"sortBy", bson.D{{"CreatedAt", -1}}},
															}},
														}},
														{"as", "quality"},
														{"cond", bson.D{
															{"$and", bson.A{
																bson.D{{"$eq", bson.A{"$$quality.server_id", "$$server._id"}}},
																bson.D{{"$eq", bson.A{"$$quality.episode_id", "$$episode._id"}}},
																bson.D{{"$ne", bson.A{"$$quality.deleted", "deleted"}}},
																bson.D{{"$ne", bson.A{"$$quality.status", 2}}},
															}},
														}},
													}},
												}}},
											}},
										}},
									}},
								}}},
							}},
						}},
					}},
				}},
			}}},

			// Final projection to organize fields
			bson.D{{"$project", bson.D{
				{"_id", 1},
				{"title", 1},
				{"name_eng", 1},
				{"description", 1},
				{"tags", 1},
				{"status", 1},
				{"image", 1},
				{"moreimage", 1},
				{"slug", 1},
				{"category", 1},
				{"genre", 1},
				{"country", 1},
				{"episode", 1},
				{"hotmovie", 1},
				{"maxquality", 1},
				{"sub", 1},
				{"trailer", 1},
				{"year", 1},
				{"season", 1},
				{"duration", 1},
				{"numofep", 1},
				{"position", 1},
				{"created_at", 1},
				{"updated_at", 1},
				{"deleted", 1},
				{"genreDetails", 1},
				{"episodeDetails", 1},
			}}},
		}

		// Thực thi pipeline
		cursor, err := movieCollection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, fmt.Errorf("Error running aggregation: %v", err)
		}
		defer cursor.Close(ctx)

		// Giải mã kết quả
		var movies []models.Movie
		if err := cursor.All(ctx, &movies); err != nil {
			return nil, fmt.Errorf("Error decoding aggregation result: %v", err)
		}
		if len(movies) == 0 {
			return nil, fmt.Errorf("Movie not found")
		}
		movie := movies[0]

		// Serialize dữ liệu để lưu vào Redis cache
		jsonData, err := json.Marshal(movie)
		if err != nil {
			log.Printf("Error serializing movie data for cache: %v", err)
		} else {
			// Lưu vào Redis với TTL 10 phút
			cacheErr := dbs.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err()
			if cacheErr != nil {
				log.Printf("Error caching movie data: %v", cacheErr)
			}
		}

		// Trả về dữ liệu phim
		return &movie, nil
	} else if err != nil {
		// Lỗi khi truy cập Redis
		return nil, fmt.Errorf("Error fetching cache: %v", err)
	}

	// Redis cache tồn tại, giải mã JSON thành kiểu models.Movie
	var movie models.Movie
	err = json.Unmarshal([]byte(cachedMovie), &movie)
	if err != nil {
		return nil, fmt.Errorf("Error decoding cached movie data: %v", err)
	}

	// Trả về dữ liệu phim từ cache
	return &movie, nil
}
