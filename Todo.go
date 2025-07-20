package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render
var todoCollection *mongo.Collection
var ctx context.Context

const (
	mongoURI       = "mongodb://127.0.0.1:27017"
	dbName         = "demo_todo"
	collectionName = "todo"
	port           = ":9000"
)

type (
	todoModel struct {
		ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
		Title     string             `bson:"title" json:"title"`
		Completed bool               `bson:"completed" json:"completed"`
		CreatedAt time.Time          `bson:"createdAt" json:"created_at"`
	}

	todo struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
)

func init() {
	rnd = renderer.New()
	ctx = context.Background()

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	checkErr(err)

	err = client.Connect(ctx)
	checkErr(err)

	// Ping to ensure connection
	err = client.Ping(ctx, nil)
	checkErr(err)

	todoCollection = client.Database(dbName).Collection(collectionName)
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	server := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("✅ Server listening on port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	<-stopChan
	log.Println("🔻 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("🛑 Server gracefully stopped")
}

func todoHandlers() http.Handler {
	r := chi.NewRouter()
	r.Get("/", fetchTodos)
	r.Post("/", createTodo)
	r.Put("/{id}", updateTodo)
	r.Delete("/{id}", deleteTodo)
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusUnprocessableEntity, err)
		return
	}

	if strings.TrimSpace(t.Title) == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "The title field is required"})
		return
	}

	newTodo := todoModel{
		ID:        primitive.NewObjectID(),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	_, err := todoCollection.InsertOne(ctx, newTodo)
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to create todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{"message": "Todo created successfully", "todo_id": newTodo.ID.Hex()})
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	cursor, err := todoCollection.Find(ctx, bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to fetch todos", "error": err})
		return
	}
	defer cursor.Close(ctx)

	var todos []todoModel
	if err := cursor.All(ctx, &todos); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Cursor decode error", "error": err})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{"data": todos})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid todo ID"})
		return
	}

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusUnprocessableEntity, err)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"title":     t.Title,
			"completed": t.Completed,
		},
	}

	_, err = todoCollection.UpdateByID(ctx, objID, update)
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to update todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo updated successfully"})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid todo ID"})
		return
	}

	_, err = todoCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to delete todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo deleted successfully"})
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
