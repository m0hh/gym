package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.requireCoach(app.registerUserHandler))
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/coach/get", app.requireCoach(app.listCoachusers))
	router.HandlerFunc(http.MethodPost, "/v1/users/card/weight", app.requireTrainee(app.updateUserWeight))
	router.HandlerFunc(http.MethodGet, "/v1/users/card/get", app.requireTrainee(app.retrieveUserCard))
	router.HandlerFunc(http.MethodGet, "/v1/users/histories/list", app.requireTrainee(app.listuserHistory))
	router.HandlerFunc(http.MethodGet, "/v1/users/coach/histories/list/:id", app.requireCoach(app.listuserHistoryCoach))

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/meals/food/add", app.requireCoachOrAdmin(app.addFood))
	router.HandlerFunc(http.MethodGet, "/v1/meals/food/get/:id", app.retriveFood)
	router.HandlerFunc(http.MethodPatch, "/v1/meals/food/update/:id", app.requireAdmin(app.updateFood))
	router.HandlerFunc(http.MethodGet, "/v1/meals/food/get", app.listFoods)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/food/delete/:id", app.requireAdmin(app.deleteFoodHandler))

	router.HandlerFunc(http.MethodPost, "/v1/meals/breakfast/add", app.requireCoachOrAdmin(app.addBreakfast))
	router.HandlerFunc(http.MethodPatch, "/v1/meals/breakfast/update/:id", app.requireAdmin(app.updateBreakfast))
	router.HandlerFunc(http.MethodGet, "/v1/meals/breakfast/get/:id", app.getBreakfastFoodById)
	router.HandlerFunc(http.MethodGet, "/v1/meals/breakfast/get", app.listBreakfasts)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/breakfast/delete/:id", app.requireAdmin(app.deleteBreakfastHandler))

	router.HandlerFunc(http.MethodPost, "/v1/meals/amsnack/add", app.requireCoachOrAdmin(app.addAmSnack))
	router.HandlerFunc(http.MethodPatch, "/v1/meals/amsnack/update/:id", app.requireAdmin(app.updateAmSnack))
	router.HandlerFunc(http.MethodGet, "/v1/meals/amsnack/get/:id", app.getAmSnackFoodById)
	router.HandlerFunc(http.MethodGet, "/v1/meals/amsnack/get", app.listAmSnacks)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/amsnack/delete/:id", app.requireAdmin(app.deleteAmSnackHandler))

	router.HandlerFunc(http.MethodPost, "/v1/meals/lunch/add", app.requireCoachOrAdmin(app.addLunch))
	router.HandlerFunc(http.MethodPatch, "/v1/meals/lunch/update/:id", app.requireAdmin(app.updateLunch))
	router.HandlerFunc(http.MethodGet, "/v1/meals/lunch/get/:id", app.getLunchById)
	router.HandlerFunc(http.MethodGet, "/v1/meals/lunch/get", app.listlunches)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/lunch/delete/:id", app.requireAdmin(app.deleteLunchHandler))

	router.HandlerFunc(http.MethodPost, "/v1/meals/pmsnack/add", app.requireCoachOrAdmin(app.addPmSnack))
	router.HandlerFunc(http.MethodPatch, "/v1/meals/pmsnack/update/:id", app.requireAdmin(app.updatePmSnack))
	router.HandlerFunc(http.MethodGet, "/v1/meals/pmsnack/get/:id", app.getPmSnackById)
	router.HandlerFunc(http.MethodGet, "/v1/meals/pmsnack/get", app.listPmsnacks)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/pmsnack/delete/:id", app.requireAdmin(app.deletePmSnackHandler))

	router.HandlerFunc(http.MethodPost, "/v1/meals/dinner/add", app.requireCoachOrAdmin(app.addDinner))
	router.HandlerFunc(http.MethodPatch, "/v1/meals/dinner/update/:id", app.requireAdmin(app.updateDinner))
	router.HandlerFunc(http.MethodGet, "/v1/meals/dinner/get/:id", app.getDinnerById)
	router.HandlerFunc(http.MethodGet, "/v1/meals/dinner/get", app.listDinners)
	router.HandlerFunc(http.MethodDelete, "/v1/meals/dinner/delete/:id", app.requireAdmin(app.deleteDinnerHandler))

	router.HandlerFunc(http.MethodPost, "/v1/plans/day/add", app.requireCoachOrAdmin(app.addDay))
	router.HandlerFunc(http.MethodGet, "/v1/plans/day/get/:id", app.requireCoachOrAdmin(app.retrieveDay))
	router.HandlerFunc(http.MethodGet, "/v1/plans/day/get", app.requireCoachOrAdmin(app.listDays))

	router.HandlerFunc(http.MethodPost, "/v1/plans/planmeal/add", app.requireCoachOrAdmin(app.addPlanMeal))
	router.HandlerFunc(http.MethodGet, "/v1/plans/planmeal/get", app.requireCoachOrAdmin(app.listPlans))
	router.HandlerFunc(http.MethodPut, "/v1/plans/planmeal/user/add", app.requireCoach(app.addPlantoUser))

	router.HandlerFunc(http.MethodPost, "/v1/exercise/name/add", app.requireCoachOrAdmin(app.addExerciseName))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/name/list", app.listExerciseNames)
	router.HandlerFunc(http.MethodGet, "/v1/exercise/name/get/:id", app.getExerciseName)
	router.HandlerFunc(http.MethodDelete, "/v1/exercise/name/delete/:id", app.requireAdmin(app.deleteExerciseName))

	router.HandlerFunc(http.MethodPost, "/v1/exercise/add", app.requireCoach(app.addExercise))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/list", app.requireCoach(app.listExercises))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/get/:id", app.requireCoach(app.getExercise))
	router.HandlerFunc(http.MethodDelete, "/v1/exercise/delete/:id", app.requireCoach(app.deleteExercise))

	router.HandlerFunc(http.MethodPost, "/v1/exercise/day/add", app.requireCoach(app.addExerciseDay))
	router.HandlerFunc(http.MethodPatch, "/v1/exercise/day/update/:id", app.requireCoach(app.updateExcerciseDay))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/day/list", app.requireCoach(app.listExerciseDays))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/day/get/:id", app.requireCoach(app.getExerciseDay))
	router.HandlerFunc(http.MethodDelete, "/v1/exercise/day/delete/:id", app.requireCoach(app.deleteExerciseDay))

	router.HandlerFunc(http.MethodPost, "/v1/exercise/plan/add", app.requireCoach(app.addExercisePlan))
	router.HandlerFunc(http.MethodPatch, "/v1/exercise/plan/update/:id", app.requireCoach(app.updateExcercisePlan))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/plan/list", app.requireCoach(app.listExercisePlans))
	router.HandlerFunc(http.MethodGet, "/v1/exercise/plan/get/:id", app.requireCoach(app.getExercisePlan))
	router.HandlerFunc(http.MethodDelete, "/v1/exercise/plan/delete/:id", app.requireCoach(app.deleteExercisePlan))
	router.HandlerFunc(http.MethodPut, "/v1/exercise/plan/assign", app.requireCoach(app.addExPlantoUser))

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
