package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/internal/handler"
	"blogwitgoofy/server/internal/middleware"
)

func Setup(cfg *config.Config) *gin.Engine {
	handler.Init(cfg)

	r := gin.Default()

	// 开发环境直接跨域放开；生产走 nginx 反代同源
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 上传图片静态服务
	r.Static("/uploads", "./uploads")

	r.Use(middleware.ApiLog())

	v1 := r.Group("/api/v1")

	// ---- 公开接口 ----
	v1.POST("/auth/login", handler.Login)

	v1.GET("/articles", handler.ListArticles)
	v1.GET("/articles/archives", handler.Archives)
	v1.GET("/articles/:id", handler.GetArticle)
	v1.GET("/articles/:id/comments", handler.ListComments)
	v1.POST("/articles/:id/comments", handler.CreateComment)
	v1.GET("/articles/:id/annotations", handler.ListAnnotations)

	v1.GET("/categories", handler.ListCategories)
	v1.GET("/tags", handler.ListTags)
	v1.GET("/links", handler.ListLinks)
	v1.GET("/pages/:slug", handler.GetPage)

	// 学习模块：书架 / 章节目录 / 章节正文 / 翻译代理 / 阅读笔记
	v1.GET("/books", handler.ListBooks)
	v1.GET("/books/:id", handler.GetBook)
	v1.GET("/chapters/:id", handler.GetChapter)
	v1.GET("/translate", handler.Translate)
	v1.GET("/chapters/:id/notes", handler.ListChapterNotes)
	v1.GET("/words", handler.ListWords)
	v1.GET("/settings", handler.GetSettings)
	v1.GET("/animes", handler.ListAnimes)
	v1.GET("/arts", handler.ListArts)
	v1.GET("/ai-tools", handler.ListAiTools)
	v1.GET("/ai-prompts", handler.ListAiPrompts)
	v1.GET("/games", handler.ListGames)
	v1.GET("/games/:id", handler.GetGame)
	v1.GET("/games/:id/posts", handler.ListGamePosts)
	v1.GET("/game-posts/:id", handler.GetGamePost)
	v1.GET("/game-posts/:id/comments", handler.ListGameComments)
	v1.POST("/game-posts/:id/comments", handler.CreateGameComment)

	// ---- 管理接口（JWT 鉴权） ----
	admin := v1.Group("/admin", middleware.JWTAuth(cfg))

	admin.GET("/dashboard", handler.Dashboard)
	admin.PUT("/password", handler.ChangePassword)

	admin.GET("/articles", handler.AdminListArticles)
	admin.POST("/articles", handler.CreateArticle)
	admin.PUT("/articles/:id", handler.UpdateArticle)
	admin.DELETE("/articles/:id", handler.DeleteArticle)

	admin.POST("/categories", handler.CreateCategory)
	admin.PUT("/categories/:id", handler.UpdateCategory)
	admin.DELETE("/categories/:id", handler.DeleteCategory)

	admin.POST("/tags", handler.CreateTag)
	admin.PUT("/tags/:id", handler.UpdateTag)
	admin.DELETE("/tags/:id", handler.DeleteTag)

	admin.GET("/comments", handler.AdminListComments)
	admin.DELETE("/comments/:id", handler.DeleteComment)

	admin.POST("/annotations", handler.CreateAnnotation)
	admin.PUT("/annotations/:id", handler.UpdateAnnotation)
	admin.DELETE("/annotations/:id", handler.DeleteAnnotation)

	admin.POST("/links", handler.CreateLink)
	admin.PUT("/links/:id", handler.UpdateLink)
	admin.DELETE("/links/:id", handler.DeleteLink)

	admin.GET("/pages", handler.AdminListPages)
	admin.POST("/pages", handler.CreatePage)
	admin.PUT("/pages/:id", handler.UpdatePage)
	admin.DELETE("/pages/:id", handler.DeletePage)

	admin.GET("/books", handler.AdminListBooks)
	admin.POST("/books", handler.CreateBook)
	admin.PUT("/books/:id", handler.UpdateBook)
	admin.DELETE("/books/:id", handler.DeleteBook)

	admin.GET("/books/:id/chapters", handler.AdminListChapters)
	admin.POST("/chapters", handler.CreateChapter)
	admin.POST("/chapters/batch", handler.BatchCreateChapters)
	admin.PUT("/chapters/:id", handler.UpdateChapter)
	admin.DELETE("/chapters/:id", handler.DeleteChapter)

	admin.POST("/chapter-notes", handler.CreateChapterNote)
	admin.PUT("/chapter-notes/:id", handler.UpdateChapterNote)
	admin.DELETE("/chapter-notes/:id", handler.DeleteChapterNote)

	admin.POST("/words", handler.AddWord)
	admin.DELETE("/words/:id", handler.DeleteWord)

	admin.PUT("/settings", handler.SaveSettings)

	admin.GET("/animes", handler.AdminListAnimes)
	admin.POST("/animes", handler.CreateAnime)
	admin.PUT("/animes/:id", handler.UpdateAnime)
	admin.DELETE("/animes/:id", handler.DeleteAnime)

	admin.GET("/arts", handler.AdminListArts)
	admin.POST("/arts", handler.CreateArt)
	admin.PUT("/arts/:id", handler.UpdateArt)
	admin.DELETE("/arts/:id", handler.DeleteArt)

	admin.GET("/ai-tools", handler.AdminListAiTools)
	admin.POST("/ai-tools", handler.CreateAiTool)
	admin.PUT("/ai-tools/:id", handler.UpdateAiTool)
	admin.DELETE("/ai-tools/:id", handler.DeleteAiTool)

	admin.GET("/ai-prompts", handler.AdminListAiPrompts)
	admin.POST("/ai-prompts", handler.CreateAiPrompt)
	admin.PUT("/ai-prompts/:id", handler.UpdateAiPrompt)
	admin.DELETE("/ai-prompts/:id", handler.DeleteAiPrompt)

	admin.GET("/apilogs", handler.AdminListApiLogs)
	admin.GET("/apilogs/stats", handler.ApiLogStats)
	admin.DELETE("/apilogs", handler.ClearApiLogs)

	admin.GET("/games", handler.AdminListGames)
	admin.GET("/game-posts", handler.AdminListGamePosts)
	admin.POST("/game-posts", handler.CreateGamePost)
	admin.PUT("/game-posts/:id", handler.UpdateGamePost)
	admin.DELETE("/game-posts/:id", handler.DeleteGamePost)
	admin.GET("/game-comments", handler.AdminListGameComments)
	admin.DELETE("/game-comments/:id", handler.DeleteGameComment)
	admin.POST("/games", handler.CreateGame)
	admin.PUT("/games/:id", handler.UpdateGame)
	admin.DELETE("/games/:id", handler.DeleteGame)

	admin.POST("/upload", handler.Upload)

	return r
}
