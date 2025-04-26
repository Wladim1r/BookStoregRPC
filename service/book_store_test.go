package service_test

import (
	"bookstoregrpc/pb"
	"bookstoregrpc/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initTestDB(t *testing.T) *gorm.DB {
	dbname := "file:testdb_" + t.Name() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&pb.Book{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	return db
}

func TestCreateAndReadBook_store(t *testing.T) {
	t.Parallel()
	db := initTestDB(t)
	store := service.NewPostgresStore(db)

	book := &pb.Book{
		Id:     t.Name(),
		Author: "test",
		Title:  "create and get",
		Price:  123,
	}

	id, err := store.CreateBook(book)
	assert.NoError(t, err)
	assert.Equal(t, book.Id, id)

	t.Run("Success Get Book", func(t *testing.T) {
		get, err := store.GetBook(id)
		assert.NoError(t, err)
		assert.Equal(t, book.Title, get.Title)
	})

	t.Run("Failed Get Book", func(t *testing.T) {
		get, err := store.GetBook("wrong")
		assert.Error(t, err)
		assert.Empty(t, get)
	})
}

func TestUpdateBook_store(t *testing.T) {
	t.Parallel()
	db := initTestDB(t)
	store := service.NewPostgresStore(db)

	book := &pb.Book{
		Id:     t.Name(),
		Author: "test",
		Title:  "create and get",
		Price:  123,
	}

	NewBook_Success := &pb.Book{
		Id:     t.Name(),
		Author: "change test",
		Title:  "get change book",
		Price:  321,
	}
	NewBook_Failed := &pb.Book{
		Id:     "not similar id",
		Author: "change test",
		Title:  "get change book",
		Price:  321,
	}

	id, err := store.CreateBook(book)
	assert.NoError(t, err)

	t.Run("Success Update", func(t *testing.T) {
		_, err := store.UpdateBook(id, NewBook_Success)
		assert.NoError(t, err)
		alterBook, err := store.GetBook(id)
		assert.NoError(t, err)
		assert.Equal(t, NewBook_Success, alterBook)
	})

	t.Run("Failed Update", func(t *testing.T) {
		alterBook, err := store.UpdateBook(id, NewBook_Failed)
		assert.Empty(t, alterBook)
		assert.Equal(t, errors.New("Book ID must be similar"), err)
	})

}

func TestDeleteBook_store(t *testing.T) {
	t.Parallel()
	db := initTestDB(t)
	store := service.NewPostgresStore(db)

	book := &pb.Book{
		Id:     t.Name(),
		Author: "test",
		Title:  "create and get",
		Price:  123,
	}

	id, err := store.CreateBook(book)
	assert.NoError(t, err)

	t.Run("Success Delete", func(t *testing.T) {
		deletedBook, err := store.DeleteBook(id)
		assert.NoError(t, err)
		assert.Equal(t, book, deletedBook)
	})
	t.Run("Failed Delete", func(t *testing.T) {
		deletedBook, err := store.DeleteBook(id)
		assert.Error(t, err)
		assert.Empty(t, deletedBook)
	})
}

func TestSearchAndGetAllBooks_store(t *testing.T) {
	t.Parallel()
	db := initTestDB(t)
	store := service.NewPostgresStore(db)

	books := []*pb.Book{
		{Id: t.Name() + "_1", Author: "case 1", Title: "test 1", Price: 123},
		{Id: t.Name() + "_2", Author: "case 1", Title: "test 2", Price: 234},
		{Id: t.Name() + "_3", Author: "case 3", Title: "test 3", Price: 345},
	}

	for _, book := range books {
		_, err := store.CreateBook(book)
		assert.NoError(t, err)
	}

	t.Run("Get All Books", func(t *testing.T) {
		books, err := store.GetBooks()
		assert.NoError(t, err)
		assert.Len(t, books, 3)
	})

	t.Run("Search Books", func(t *testing.T) {
		filter := &pb.Filter{
			Author: "case 1",
			Price:  123,
		}

		books, err := store.SearchBook(filter)
		assert.NoError(t, err)
		assert.Len(t, books, 1)
	})

	t.Run("Search Books Empty response", func(t *testing.T) {
		filter := &pb.Filter{
			Author: "case 1",
			Price:  300,
		}

		books, err := store.SearchBook(filter)
		assert.NoError(t, err)
		assert.Empty(t, books)
	})
}
