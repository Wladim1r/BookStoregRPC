package service

import (
	"bookstoregrpc/pb"
	"errors"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

type BookStote interface {
	GetBook(string) (*pb.Book, error)
	GetBooks() ([]*pb.Book, error)
	CreateBook(*pb.Book) (string, error)
	UpdateBook(string, *pb.Book) (*pb.Book, error)
	DeleteBook(string) (*pb.Book, error)
	SearchBook(*pb.Filter) ([]*pb.Book, error)
}

type PostgresStore struct {
	mu sync.RWMutex
	db *gorm.DB
}

func NewPostgresStore(db *gorm.DB) BookStote {
	return &PostgresStore{db: db}
}

func (ps *PostgresStore) GetBook(id string) (*pb.Book, error) {
	log.Println("GETBOOK receive request")
	var book *pb.Book

	ps.mu.RLock()
	err := ps.db.Where("id = ?", id).First(&book).Error
	ps.mu.RUnlock()

	return book, err
}

func (ps *PostgresStore) GetBooks() ([]*pb.Book, error) {
	log.Println("GETBOOKS receive request")
	var books []*pb.Book

	ps.mu.RLock()
	err := ps.db.Find(&books).Error
	ps.mu.RUnlock()

	return books, err
}

func (ps *PostgresStore) CreateBook(book *pb.Book) (string, error) {
	log.Println("CREATEBOOK receive request")
	ps.mu.Lock()
	err := ps.db.Create(&book).Error
	ps.mu.Unlock()

	return book.Id, err
}

func (ps *PostgresStore) UpdateBook(id string, newBook *pb.Book) (*pb.Book, error) {
	log.Println("UPDATEBOOK receive request")
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var book *pb.Book
	err := ps.db.Where("id = ?", id).First(&book).Error
	if err != nil {
		return nil, err
	}

	if book.Id != newBook.Id {
		return nil, errors.New("Book ID must be similar")
	}

	book.Author = newBook.Author
	book.Title = newBook.Title
	book.Price = newBook.Price

	err = ps.db.Model(&pb.Book{}).Where("id = ?", id).Select("Title", "Author", "Price").Updates(book).Error
	if err != nil {
		return nil, err
	}

	return book, err
}

func (ps *PostgresStore) DeleteBook(id string) (*pb.Book, error) {
	log.Println("DELETEBOOK receive request")
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var book *pb.Book
	err := ps.db.Where("id = ?", id).First(&book).Error
	if err != nil {
		return nil, err
	}

	err = ps.db.Unscoped().Where("id = ?", id).Delete(&pb.Book{}).Error
	if err != nil {
		return nil, err
	}

	return book, err
}

func (ps *PostgresStore) SearchBook(filter *pb.Filter) ([]*pb.Book, error) {
	log.Println("SEARCHBOOK receive request")
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	query := ps.db.Model(&pb.Book{})

	if filter.GetAuthor() != "" {
		query = query.Where("author = ?", filter.GetAuthor())
	}
	if filter.GetPrice() > 0 {
		query = query.Where("price > ?", filter.GetPrice())
	}

	var books []*pb.Book
	if err := query.Find(&books).Error; err != nil {
		return nil, fmt.Errorf("failed to search books: %w", err)
	}

	return books, nil
}
