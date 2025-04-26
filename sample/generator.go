package sample

import (
	"bookstoregrpc/pb"
	"math/rand/v2"

	"github.com/google/uuid"
)

func NewBook() *pb.Book {
	author := RandomAuthor()
	return &pb.Book{
		Id:     RandomID(),
		Author: author,
		Title:  RandomTitle(author),
		Price:  rand.Int32N(1000) + 200,
	}
}

func RandomID() string {
	return uuid.New().String()
}

func RandomAuthor() string {
	authors := []string{"А. С. Пушкин", "Ф. М. Достоевский", "Л. Н. Толстой", "Н. В. Гоголь", "А. П. Чехов"}
	return authors[rand.IntN(len(authors))]
}

func RandomTitle(author string) string {
	n := 3

	switch author {
	case "А. С. Пушкин":
		titles := []string{"Евгений Онегин", "Капитанская дочка", "Сказка о царе Салтане"}
		return titles[rand.IntN(n)]
	case "Ф. М. Достоевский":
		titles := []string{"Преступление и наказание", "Идиот", "Братья Карамазовы"}
		return titles[rand.IntN(n)]
	case "Л. Н. Толстой":
		titles := []string{"Война и мир", "Анна Каренина", "Смерть Ивана Ильича"}
		return titles[rand.IntN(n)]
	case "Н. В. Гоголь":
		titles := []string{"Мёртвые души", "Ревизор", "Вечера на хуторе близ Диканьки"}
		return titles[rand.IntN(n)]
	default:
		titles := []string{"Вишнёвый сад", "Рассказы", "Чайка"}
		return titles[rand.IntN(n)]
	}
}
