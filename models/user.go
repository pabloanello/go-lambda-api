package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (ur *UserRequest) Validate(isUpdate bool) error {
	if !isUpdate {
		if ur.Name == "" {
			return errors.New("name is required")
		}
		if ur.Email == "" {
			return errors.New("email is required")
		}
	} else if ur.Name == "" && ur.Email == "" {
		return errors.New("no fields to update")
	}

	return nil
}

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(user User) (User, error)
	GetUserByID(id string) (User, error)
	GetAllUsers() []User
	UpdateUser(user User) (User, error)
	DeleteUser(id string) error
}

// inMemoryUserRepository implements UserRepository using an in-memory map.
type inMemoryUserRepository struct {
	users map[string]User
}

// globalInMemoryUserRepository is the singleton instance for in-memory user storage.
var globalInMemoryUserRepository = &inMemoryUserRepository{users: make(map[string]User)}

// NewInMemoryUserRepository creates a new instance of inMemoryUserRepository.
// nolint: ireturn
func NewInMemoryUserRepository() UserRepository {
	return globalInMemoryUserRepository
}

// ClearInMemoryUsers clears the in-memory user store for testing.
func ClearInMemoryUsers() {
	globalInMemoryUserRepository.users = make(map[string]User)
}

// ClearUsers clears the in-memory user store for testing.
func (r *inMemoryUserRepository) ClearUsers() {
	r.users = make(map[string]User)
}

func (r *inMemoryUserRepository) GetUserByID(id string) (User, error) {
	user, exists := r.users[id]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

func (r *inMemoryUserRepository) GetAllUsers() []User {
	userList := make([]User, 0, len(r.users))
	for _, user := range r.users {
		userList = append(userList, user)
	}

	return userList
}

func (r *inMemoryUserRepository) CreateUser(user User) (User, error) {
	r.users[user.ID] = user

	return user, nil
}

func (r *inMemoryUserRepository) UpdateUser(user User) (User, error) {
	_, exists := r.users[user.ID]
	if !exists {
		return User{}, errors.New("user not found")
	}
	r.users[user.ID] = user

	return user, nil
}

func (r *inMemoryUserRepository) DeleteUser(id string) error {
	_, exists := r.users[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(r.users, id)

	return nil
}

// dynamoDBUserRepository implements UserRepository for DynamoDB.
type dynamoDBUserRepository struct {
	db        dynamodbiface.DynamoDBAPI
	tableName string
}

// NewDynamoDBUserRepository creates a new instance of dynamoDBUserRepository.
func NewDynamoDBUserRepository(db dynamodbiface.DynamoDBAPI, tableName string) UserRepository {
	return &dynamoDBUserRepository{db: db, tableName: tableName}
}

// CreateUser inserts a new user into DynamoDB.
func (r *dynamoDBUserRepository) CreateUser(user User) (User, error) {
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return User{}, fmt.Errorf("failed to marshal user: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.tableName),
	}

	_, err = r.db.PutItem(input)
	if err != nil {
		return User{}, fmt.Errorf("failed to put item to DynamoDB: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user from DynamoDB by ID.
func (r *dynamoDBUserRepository) GetUserByID(id string) (User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(r.tableName),
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		return User{}, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return User{}, errors.New("user not found")
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return User{}, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return user, nil
}

// GetAllUsers retrieves all users from DynamoDB.
func (r *dynamoDBUserRepository) GetAllUsers() []User {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	result, err := r.db.Scan(input)
	if err != nil {
		// Log the error, but return an empty list as per the interface signature
		fmt.Printf("failed to scan items from DynamoDB: %v\n", err)
		return []User{}
	}

	var users []User
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &users)
	if err != nil {
		// Log the error, but return an empty list
		fmt.Printf("failed to unmarshal scan items: %v\n", err)
		return []User{}
	}

	return users
}

// UpdateUser updates an existing user in DynamoDB.
func (r *dynamoDBUserRepository) UpdateUser(user User) (User, error) {
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return User{}, fmt.Errorf("failed to marshal user: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(r.tableName),
	}

	_, err = r.db.PutItem(input)
	if err != nil {
		return User{}, fmt.Errorf("failed to update item in DynamoDB: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user from DynamoDB by ID.
func (r *dynamoDBUserRepository) DeleteUser(id string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(r.tableName),
	}

	_, err := r.db.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete item from DynamoDB: %w", err)
	}

	return nil
}
