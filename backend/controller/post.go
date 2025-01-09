package controller

import (
	"TalkSphere/dao/mysql"
	"TalkSphere/models"
	"TalkSphere/pkg/upload"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreatePostRequest 创建帖子请求参数
type CreatePostRequest struct {
	Title    string   `json:"title" binding:"required"`
	Content  string   `json:"content" binding:"required"`
	BoardID  int64    `json:"board_id"`
	Tags     []string `json:"tags"`
	ImageIDs []int64  `json:"image_ids"` // 已上传图片的ID列表
}

// UpdatePostRequest 更新帖子请求参数
type UpdatePostRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	BoardID  int64    `json:"board_id"`
	Tags     []string `json:"tags"`
	ImageIDs []int64  `json:"image_ids"`
}

// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	userIDInterface, exists := c.Get(CtxtUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 正确的类型断言
	userID, ok := userIDInterface.(int64)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}

	// 创建帖子
	post := &models.Post{
		Title:    req.Title,
		Content:  req.Content,
		BoardID:  &req.BoardID,
		AuthorID: &userID,
	}

	// 开启事务
	tx := mysql.DB.Begin()
	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		zap.L().Error("create post failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 处理标签
	if len(req.Tags) > 0 {
		var tags []models.Tag
		for _, tagName := range req.Tags {
			var tag models.Tag
			// 查找或创建标签
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				zap.L().Error("create post failed", zap.Error(err))
				ResponseError(c, CodeServerBusy)
				return
			}
			tags = append(tags, tag)
		}
		if err := tx.Model(post).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			zap.L().Error("create post failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}
	}
	//前端传入的
	// 处理图片
	if len(req.ImageIDs) > 0 {
		// 验证图片所属权并更新关联
		var images []models.PostImage
		if err := tx.Where("id IN ? AND user_id = ? AND status = 1 AND post_id = 0",
			req.ImageIDs, userID).Find(&images).Error; err != nil {
			tx.Rollback()
			zap.L().Error("find images failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}

		// 确保所有图片都存在且属于当前用户
		if len(images) != len(req.ImageIDs) {
			tx.Rollback()
			ResponseError(c, CodeInvalidParam)
			return
		}

		// 更新图片关联
		for i, img := range images {
			if err := tx.Model(&img).Updates(map[string]interface{}{
				"post_id":    post.ID,
				"sort_order": i,
			}).Error; err != nil {
				tx.Rollback()
				zap.L().Error("update image failed", zap.Error(err))
				ResponseError(c, CodeServerBusy)
				return
			}
		}
	}

	tx.Commit()
	ResponseSuccess(c, gin.H{
		"post_id": post.ID,
	})
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	// 1. 参数验证
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("parse post id failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 2. 查询帖子
	var post models.Post
	result := mysql.DB.Preload("Tags").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", 1).Order("sort_order")
		}).
		Where("status != ?", -1). // 不查询已删除的帖子
		First(&post, postID)

	// 3. 错误处理
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 帖子不存在
			zap.L().Info("post not found", zap.Int64("post_id", postID))
			ResponseError(c, CodePostNotExist)
			return
		}
		// 数据库错误
		zap.L().Error("query post failed",
			zap.Int64("post_id", postID),
			zap.Error(result.Error))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 4. 构建响应数据
	type PostResponse struct {
		ID            int64        `json:"id"`
		Title         string       `json:"title"`
		Content       string       `json:"content"`
		BoardID       *int64       `json:"board_id"`
		AuthorID      *int64       `json:"author_id"`
		ViewCount     int          `json:"view_count"`
		LikeCount     int          `json:"like_count"`
		FavoriteCount int          `json:"favorite_count"`
		CommentCount  int          `json:"comment_count"`
		CreatedAt     time.Time    `json:"created_at"`
		UpdatedAt     time.Time    `json:"updated_at"`
		Tags          []models.Tag `json:"tags"`
		ImageURLs     []string     `json:"image_urls"`
	}
	response := PostResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		BoardID:       post.BoardID,
		AuthorID:      post.AuthorID,
		ViewCount:     post.ViewCount,
		LikeCount:     post.LikeCount,
		FavoriteCount: post.FavoriteCount,
		CommentCount:  post.CommentCount,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Tags:          post.Tags,
		ImageURLs:     make([]string, 0, len(post.Images)),
	}

	// 5. 提取图片URL
	for _, img := range post.Images {
		response.ImageURLs = append(response.ImageURLs, img.ImageURL)
	}

	// 6. 异步更新浏览量
	go func() {
		if err := mysql.DB.Model(&post).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).
			Error; err != nil {
			zap.L().Error("update view count failed",
				zap.Int64("post_id", postID),
				zap.Error(err))
		}
	}()

	fmt.Println(response)

	ResponseSuccess(c, response)
}

// DeletePost 删除帖子
func DeletePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	userIDInterface, exists := c.Get(CtxtUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 正确的类型断言
	userID, ok := userIDInterface.(int64)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}

	var post models.Post
	if err := mysql.DB.First(&post, postID).Error; err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 检查是否是帖子作者
	if *post.AuthorID != userID {
		ResponseError(c, CodeNoPermision)
		return
	}

	// 软删除帖子
	if err := mysql.DB.Model(&post).Update("status", -1).Error; err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	userIDInterface, exists := c.Get(CtxtUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 正确的类型断言
	userID, ok := userIDInterface.(int64)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}
	var post models.Post
	if err := mysql.DB.First(&post, postID).Error; err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 检查是否是帖子作者
	if *post.AuthorID != userID {
		ResponseError(c, CodeNoPermision)
		return
	}

	// 开启事务
	tx := mysql.DB.Begin()

	// 更新帖子基本信息
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.BoardID != 0 {
		updates["board_id"] = req.BoardID
	}

	if err := tx.Model(&post).Updates(updates).Error; err != nil {
		tx.Rollback()
		ResponseError(c, CodeServerBusy)
		return
	}

	// 更新标签
	if len(req.Tags) > 0 {
		var tags []models.Tag
		for _, tagName := range req.Tags {
			var tag models.Tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				tx.Rollback()
				ResponseError(c, CodeServerBusy)
				return
			}
			tags = append(tags, tag)
		}
		if err := tx.Model(&post).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			ResponseError(c, CodeServerBusy)
			return
		}
	}

	// 更新图片
	if len(req.ImageIDs) > 0 {
		// 删除原有图片关联
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostImage{}).Error; err != nil {
			tx.Rollback()
			ResponseError(c, CodeServerBusy)
			return
		}

		// 添加新的图片关联
		for i, imageID := range req.ImageIDs {
			postImage := models.PostImage{
				PostID:    postID,
				ID:        imageID,
				SortOrder: i,
			}
			if err := tx.Create(&postImage).Error; err != nil {
				tx.Rollback()
				ResponseError(c, CodeServerBusy)
				return
			}
		}
	}

	tx.Commit()
	ResponseSuccess(c, nil)
}

// GetUserPosts 获取用户的帖子列表
func GetUserPosts(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	page, size := getPageInfo(c)
	var posts []models.Post
	var total int64

	db := mysql.DB.Model(&models.Post{}).Where("author_id = ? AND status != -1", userID)
	db.Count(&total)
	if err := db.Preload("Tags").Offset(int((page - 1) * size)).Limit(int(size)).Find(&posts).Error; err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"posts": posts,
		"total": total,
	})
}

// GetBoardPosts 获取板块下的帖子列表
func GetBoardPosts(c *gin.Context) {
	boardIDStr := c.Param("board_id")
	boardID, err := strconv.ParseInt(boardIDStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	page, size := getPageInfo(c)
	var posts []models.Post
	var total int64

	// 修改查询,使用正确的字段名 avatar_url
	db := mysql.DB.Model(&models.Post{}).
		Joins("LEFT JOIN users ON posts.author_id = users.id").
		Where("posts.board_id = ? AND posts.status != -1", boardID)

	db.Count(&total)

	if err := db.Preload("Author").
		Preload("Tags").
		Select("posts.*, users.username as author_username, users.avatar_url as author_avatar_url").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Find(&posts).Error; err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"posts": posts,
		"total": total,
	})
}

// UploadPostImage 上传帖子图片
func UploadPostImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	userIDInterface, exists := c.Get(CtxtUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	userID, ok := userIDInterface.(int64)
	if !ok {
		ResponseError(c, CodeServerBusy)
		return
	}

	imageURL, err := upload.SaveImageToOSS(file, "post_images", userID)
	if err != nil {
		if err.Error() == "文件大小超过限制" || err.Error() == "不支持的文件类型" {
			ResponseError(c, CodeInvalidParam)
		} else {
			ResponseError(c, CodeServerBusy)
		}
		return
	}

	postImage := &models.PostImage{
		UserID:    userID,
		ImageURL:  imageURL,
		Status:    1,
		SortOrder: 0,
	}

	if err := mysql.DB.Create(postImage).Error; err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"image_id":  postImage.ID,
		"image_url": imageURL,
	})
}
