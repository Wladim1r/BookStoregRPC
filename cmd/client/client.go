package main

import (
	"bookstoregrpc/pb"
	"bookstoregrpc/sample"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	conn, err := grpc.NewClient("0.0.0.0:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	bookClient := pb.NewBookServiceClient(conn)

	n := 10
	ids := make([]string, n)

	fmt.Printf("---------------\n| CREATE BOOK |\n---------------\n\n")
	for i := range n {
		ids[i] = create(bookClient)
	}

	fmt.Printf("------------------\n| READ ALL BOOKS |\n------------------\n\n")
	getAll(bookClient)

	fmt.Printf("-----------------\n| READ ONE BOOK |\n-----------------\n\n")
	for range 2 {
		readOne(ids[rand.IntN(n)], bookClient)
	}

	fmt.Printf("---------------\n| UPDATE BOOK |\n---------------\n\n")

	newBooks := []*pb.UpdateBookRequest{
		{
			Id: ids[2],
			Book: &pb.Book{
				Id:     ids[2],
				Author: "Тейва Харшани",
				Title:  "100 ошибок Go и как их избежать",
				Price:  1200,
			},
		},
		{
			Id: ids[4],
			Book: &pb.Book{
				Id:     ids[4],
				Author: "Дэниел Джей Барретт",
				Title:  "Linux карманный справочник (Четвертое издание)",
				Price:  1899,
			},
		},
	}

	for _, newBook := range newBooks {
		updateBook(newBook, bookClient)
	}

	fmt.Printf("-----------------\n| READ NEW BOOK |\n-----------------\n\n")
	readOne(ids[2], bookClient)
	readOne(ids[4], bookClient)

	fmt.Printf("---------------\n| DELETE BOOK |\n---------------\n\n")
	deleteBook(ids[rand.IntN(n)], bookClient)

	fmt.Printf("---------------\n| SEARCH BOOK |\n---------------\n\n")
	req1 := &pb.SearchBookRequest{
		Filter: &pb.Filter{
			Author: "Ф. М. Достоевский",
			Price:  300,
		},
	}

	req2 := &pb.SearchBookRequest{
		Filter: &pb.Filter{
			Author: "Л. Н. Толстой",
			Price:  500,
		},
	}

	searchBook(req1, bookClient)
	searchBook(req2, bookClient)
}

func searchBook(req *pb.SearchBookRequest, bookClient pb.BookServiceClient) {
	ctx := context.Background()

	stream, err := bookClient.SearchBook(ctx, req)
	if err != nil {
		log.Fatal("Could not get stream of books ", err)
	}

	log.Printf("BOOKS WITH AUTHOR NAME %s AND PRICE HIGH THEN %d\n\n", req.GetFilter().GetAuthor(), req.GetFilter().GetPrice())
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Print("The list of books is over\n")
			break
		}
		if err != nil {
			log.Fatal("Could not found book ", err)
		}

		book := res.GetBook()

		log.Printf("INFO about book with ID: %s\n", book.Id)
		log.Printf("    + author: %s\n", book.Author)
		log.Printf("    + title : %s\n", book.Title)
		log.Printf("    + price : %d\n\n", book.Price)
		time.Sleep(700 * time.Millisecond)
	}
}

func deleteBook(id string, bookClient pb.BookServiceClient) {
	ctx := context.Background()

	req := &pb.DeleteBookRequest{
		Id: id,
	}

	res, err := bookClient.DeleteBook(ctx, req)
	if err != nil {
		log.Fatal("Could not delete book ", err)
	}

	deletedBook := res.GetBook()
	log.Printf("Book was successfuly deleted\n")
	log.Printf("INFO aboud DELETED book with ID: %s\n", deletedBook.Id)
	log.Printf("    + author: %s\n", deletedBook.Author)
	log.Printf("    + title : %s\n", deletedBook.Title)
	log.Printf("    + price : %d\n\n", deletedBook.Price)
	time.Sleep(2 * time.Second)
}

func updateBook(req *pb.UpdateBookRequest, bookClient pb.BookServiceClient) {
	ctx := context.Background()

	reqOldBook, err := bookClient.ReadBook(ctx, &pb.ReadBookRequest{
		Id: req.Id,
	})
	if err != nil {
		log.Fatal("Could not read book ", err)
	}
	oldBook := reqOldBook.GetBook()

	res, err := bookClient.UpdateBook(ctx, req)
	if err != nil {
		log.Fatal("Could not update book ", err)
	}

	book := res.GetBook()

	log.Print("Book was successfuly updated\n")
	log.Printf("INFO about NEW book with ID: %s\n(- means was, + means be)\n", book.Id)
	log.Printf("    - author: %s\n", oldBook.Author)
	log.Printf("    + author: %s\n|\n", book.Author)
	log.Printf("    - title : %s\n", oldBook.Title)
	log.Printf("    + title : %s\n|\n", book.Title)
	log.Printf("    - price : %d\n", oldBook.Price)
	log.Printf("    + price : %d\n\n", book.Price)
	time.Sleep(300 * time.Millisecond)
}

func readOne(id string, bookClient pb.BookServiceClient) {
	ctx := context.Background()
	req := &pb.ReadBookRequest{
		Id: id,
	}

	res, err := bookClient.ReadBook(ctx, req)
	if err != nil {
		log.Fatal("Could not read book ", err)
	}

	book := res.GetBook()

	log.Printf("INFO about book with ID: %s\n", book.Id)
	log.Printf("    + author: %s\n", book.Author)
	log.Printf("    + title : %s\n", book.Title)
	log.Printf("    + price : %d\n\n", book.Price)
	time.Sleep(500 * time.Millisecond)
}

func getAll(bookClient pb.BookServiceClient) {
	ctx := context.Background()

	stream, err := bookClient.ReadBooks(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatal("Could not get stream of books ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Print("The list of books is over")
			break
		}
		if err != nil {
			log.Fatal("Could not read book ", err)
		}

		book := res.GetBook()

		log.Printf("INFO about book with ID: %s\n", book.Id)
		log.Printf("    + author: %s\n", book.Author)
		log.Printf("    + title : %s\n", book.Title)
		log.Printf("    + price : %d\n\n", book.Price)
		time.Sleep(1 * time.Second)
	}
}

func create(bookClient pb.BookServiceClient) string {
	ctx := context.Background()
	book := sample.NewBook()

	req := &pb.CreateBookRequest{
		Book: book,
	}

	id, err := bookClient.CreateBook(ctx, req)
	if err != nil {
		log.Fatal("Could not create a new book ", err)
	}

	log.Printf("Book successfuly created with ID %s", id)
	time.Sleep(1500 * time.Millisecond)

	return book.Id
}
