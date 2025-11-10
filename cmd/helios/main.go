package main

import (
	"fmt"
	pb "helios/generated/github.com/protocolbuffers/protobuf/examples/go/tutorialpb"
	"helios/internal/test"

	"google.golang.org/protobuf/proto"
)

func main() {
	test.HelloWorld()

	person := &pb.Person{
		Name:  "John Doe",
		Id:    1234567890,
		Email: "john.doe@example.com",
	}

	fmt.Println(person.String())

	data, err := proto.Marshal(person)
	if err != nil {
		fmt.Printf("Failed to marshal person: %v", err)
	}

	fmt.Println(data)
}
