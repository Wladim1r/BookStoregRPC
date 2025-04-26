package service

import (
	"bookstoregrpc/pb"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type BookServer struct {
	Store BookStote
	pb.UnimplementedBookServiceServer
}

func NewBookServer(store BookStote) *BookServer {
	return &BookServer{Store: store}
}

func (bs *BookServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	book := req.GetBook()

	id, err := bs.Store.CreateBook(book)
	if err != nil {
		return nil, err
	}

	res := &pb.CreateBookResponse{
		Id: id,
	}

	return res, err
}

func (bs *BookServer) ReadBook(ctx context.Context, req *pb.ReadBookRequest) (*pb.ReadBookResponse, error) {
	id := req.GetId()

	book, err := bs.Store.GetBook(id)
	if err != nil {
		return nil, err
	}

	res := &pb.ReadBookResponse{
		Book: book,
	}

	return res, err
}

func (bs *BookServer) ReadBooks(_ *emptypb.Empty, stream pb.BookService_ReadBooksServer) error {
	books, err := bs.Store.GetBooks()
	if err != nil {
		return err
	}

	for _, book := range books {
		res := &pb.ReadBooksResponse{
			Book: book,
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}
	}

	return err
}

func (bs *BookServer) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	id := req.GetId()
	newBook := req.GetBook()

	book, err := bs.Store.UpdateBook(id, newBook)
	if err != nil {
		return nil, err
	}

	res := &pb.UpdateBookResponse{
		Book: book,
	}

	return res, err
}

func (bs *BookServer) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	id := req.GetId()

	book, err := bs.Store.DeleteBook(id)
	if err != nil {
		return nil, err
	}

	res := &pb.DeleteBookResponse{
		Book: book,
	}

	return res, err
}

func (bs *BookServer) SearchBook(req *pb.SearchBookRequest, stream pb.BookService_SearchBookServer) error {
	filter := req.GetFilter()

	books, err := bs.Store.SearchBook(filter)
	if err != nil {
		return err
	}

	for _, book := range books {
		res := &pb.SearchBookResponse{
			Book: book,
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}
	}

	return err
}
