package routes

import (
	"github.com/Project-DSView/backend/go/internal/api/handler"
	"github.com/gofiber/fiber/v2"
)

// SetupCourseWeekRoutes sets up course week related routes
func SetupCourseWeekRoutes(app *fiber.App, courseWeekHandler *handler.CourseWeekHandler) {
	api := app.Group("/api")

	courseWeeks := api.Group("/courses/:courseId/weeks")
	{
		// GET /api/courses/:courseId/weeks - Get all course weeks
		courseWeeks.Get("/", courseWeekHandler.GetCourseWeeks)

		// POST /api/courses/:courseId/weeks - Create a new course week
		courseWeeks.Post("/", courseWeekHandler.CreateCourseWeek)

		// GET /api/courses/:courseId/weeks/:weekNumber - Get specific course week
		courseWeeks.Get("/:weekNumber", courseWeekHandler.GetCourseWeek)

		// PUT /api/courses/:courseId/weeks/:weekNumber - Update course week
		courseWeeks.Put("/:weekNumber", courseWeekHandler.UpdateCourseWeek)

		// DELETE /api/courses/:courseId/weeks/:weekNumber - Delete course week
		courseWeeks.Delete("/:weekNumber", courseWeekHandler.DeleteCourseWeek)

		// GET /api/courses/:courseId/weeks/:weekNumber/title - Get week title
		courseWeeks.Get("/:weekNumber/title", courseWeekHandler.GetWeekTitle)
	}
}

// SetupCourseContentRoutes sets up course content grouped by week routes
func SetupCourseContentRoutes(app *fiber.App, courseContentHandler *handler.CourseContentHandler) {
	api := app.Group("/api")

	courses := api.Group("/courses/:courseId")
	{
		// GET /api/courses/:courseId/content - Get all course content grouped by week
		courses.Get("/content", courseContentHandler.GetCourseContentByWeek)

		// GET /api/courses/:courseId/content/:weekNumber - Get content for specific week
		courses.Get("/content/:weekNumber", courseContentHandler.GetWeekContent)
	}
}
