package service_test

import (
	"bookstoregrpc/pb"
	"bookstoregrpc/service"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

func startTestServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	// init DB
	db := initTestDB(t)
	store := service.NewPostgresStore(db)
	server := service.NewBookServer(store)

	// create listener with buffer
	listener := bufconn.Listen(1024 * 1024)

	// create grpc server
	s := grpc.NewServer()
	pb.RegisterBookServiceServer(s, server)

	serverErr := make(chan error)

	// launch server
	go func() {
		if err := s.Serve(listener); err != nil {
			t.Logf("Could not start server %v", err)
			serverErr <- err
			close(serverErr)
		}
	}()

	select {
	case err := <-serverErr:
		t.Fatalf("Server failed to start %v", err)
		return nil, nil
	case <-time.After(300 * time.Millisecond):
		return s, listener
	}
}

type testClient struct {
	client pb.BookServiceClient
	server *grpc.Server
	conn   *grpc.ClientConn
}

func (tc *testClient) Close() {
	tc.server.Stop()
	tc.conn.Close()
}

func initClient(t *testing.T) *testClient {
	server, listener := startTestServer(t)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)

	return &testClient{
		client: pb.NewBookServiceClient(conn),
		server: server,
		conn:   conn,
	}
}

func TestCreateAndReadBook_server(t *testing.T) {
	ctx := context.Background()
	clientSTRUCT := initClient(t)
	defer clientSTRUCT.Close()

	client := clientSTRUCT.client

	book := &pb.Book{
		Id:     t.Name(),
		Author: "case 1",
		Title:  "test",
		Price:  111,
	}

	resCreate, err := client.CreateBook(ctx, &pb.CreateBookRequest{Book: book})
	assert.NoError(t, err)
	assert.Equal(t, book.Id, resCreate.GetId())

	resRead, err := client.ReadBook(ctx, &pb.ReadBookRequest{Id: resCreate.GetId()})
	assert.NoError(t, err)
	getBook := resRead.GetBook()
	assert.Equal(t, book.Id, getBook.Id)
	assert.Equal(t, book.Author, getBook.Author)
	assert.Equal(t, book.Title, getBook.Title)
	assert.Equal(t, book.Price, getBook.Price)
}

func TestUpdateBook_server(t *testing.T) {
	ctx := context.Background()
	clientSTRUCT := initClient(t)
	defer clientSTRUCT.Close()

	client := clientSTRUCT.client

	book := &pb.Book{
		Id:     t.Name(),
		Author: "case 1",
		Title:  "test",
		Price:  123,
	}
	res, err := client.CreateBook(ctx, &pb.CreateBookRequest{Book: book})
	assert.NoError(t, err)
	id := res.GetId()

	t.Run("Success Update", func(t *testing.T) {
		changeBook_OK := &pb.Book{
			Id:     id,
			Author: "case 2",
			Title:  "--test",
			Price:  321,
		}

		_, err := client.UpdateBook(ctx, &pb.UpdateBookRequest{
			Id:   id,
			Book: changeBook_OK,
		})
		assert.NoError(t, err)

		alterBookRES, err := client.ReadBook(ctx, &pb.ReadBookRequest{Id: id})
		assert.NoError(t, err)

		alterBook := alterBookRES.GetBook()
		assert.Equal(t, changeBook_OK.Author, alterBook.Author)
		assert.Equal(t, changeBook_OK.Title, alterBook.Title)
		assert.Equal(t, changeBook_OK.Price, alterBook.Price)
	})

	t.Run("Failed Update", func(t *testing.T) {
		changeBook_ERR := &pb.Book{
			Id:     t.Name() + "wrong",
			Author: "case 2",
			Title:  "--test",
			Price:  321,
		}

		_, err := client.UpdateBook(ctx, &pb.UpdateBookRequest{
			Id:   id,
			Book: changeBook_ERR,
		})
		status, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, "Book ID must be similar", status.Message())
	})

}

func TestDeleteBook_server(t *testing.T) {
	ctx := context.Background()
	clientSTRUCT := initClient(t)
	defer clientSTRUCT.Close()

	client := clientSTRUCT.client

	book := &pb.Book{
		Id:     t.Name(),
		Author: "test",
		Title:  "test",
		Price:  98,
	}
	res, err := client.CreateBook(ctx, &pb.CreateBookRequest{Book: book})
	assert.NoError(t, err)

	t.Run("Success Delete", func(t *testing.T) {
		deletedBookRES, err := client.DeleteBook(ctx, &pb.DeleteBookRequest{Id: res.GetId()})
		assert.NoError(t, err)
		deletedBook := deletedBookRES.GetBook()
		assert.Equal(t, book.Author, deletedBook.Author)
		assert.Equal(t, book.Title, deletedBook.Title)
		assert.Equal(t, book.Price, deletedBook.Price)

		_, err = client.ReadBook(ctx, &pb.ReadBookRequest{Id: res.GetId()})
		assert.Error(t, err)
	})
}

func TestSearchAndGetAllBooks_server(t *testing.T) {
	ctx := context.Background()
	clientSTRUCT := initClient(t)
	defer clientSTRUCT.Close()

	client := clientSTRUCT.client

	books := []*pb.Book{
		{Id: "1", Author: "case1", Title: "abc", Price: 130},
		{Id: "2", Author: "case1", Title: "def", Price: 23},
		{Id: "3", Author: "otherCase", Title: "xyz", Price: 138},
	}

	for _, book := range books {
		_, err := client.CreateBook(ctx, &pb.CreateBookRequest{Book: book})
		assert.NoError(t, err)
	}

	t.Run("Search Books", func(t *testing.T) {
		filter := &pb.Filter{
			Author: "case1",
			Price:  35,
		}

		res, err := client.SearchBook(ctx, &pb.SearchBookRequest{Filter: filter})
		assert.NoError(t, err)

		var i int
		for {
			_, err := res.Recv()
			if err == io.EOF {
				break
			}
			i++
			assert.NoError(t, err)
		}

		assert.Equal(t, 1, i)
	})

	t.Run("Get All Books", func(t *testing.T) {
		res, err := client.ReadBooks(ctx, &emptypb.Empty{})
		assert.NoError(t, err)

		var i int
		for {
			_, err := res.Recv()
			if err == io.EOF {
				break
			}
			i++
			assert.NoError(t, err)
		}

		assert.Equal(t, 3, i)
	})
}
