package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"imageboard/internal/models"
)

type Repository interface {
	// Topics
	CreateTopic(ctx context.Context, topic *models.Topic) error
	GetTopic(ctx context.Context, id primitive.ObjectID) (*models.Topic, error)
	ListTopics(ctx context.Context, limit, offset int) ([]models.Topic, error)
	UpdateTopicTimestamp(ctx context.Context, id primitive.ObjectID) error

	// Messages
	CreateMessage(ctx context.Context, message *models.Message) error
	GetMessagesByTopic(ctx context.Context, topicID primitive.ObjectID, limit, offset int) ([]models.Message, error)

	Close() error
}

type MongoRepository struct {
	client   *mongo.Client
	db       *mongo.Database
	topics   *mongo.Collection
	messages *mongo.Collection
}

func NewMongoRepository(uri, dbName string) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	// Create indexes
	topicsCol := db.Collection("topics")
	messagesCol := db.Collection("messages")

	// Index for topics by creation time (descending for recent first)
	_, err = topicsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdAt", Value: -1}},
	})
	if err != nil {
		return nil, err
	}

	// Index for messages by topic and creation time
	_, err = messagesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "topicId", Value: 1},
			{Key: "createdAt", Value: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	return &MongoRepository{
		client:   client,
		db:       db,
		topics:   topicsCol,
		messages: messagesCol,
	}, nil
}

func (r *MongoRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

func (r *MongoRepository) CreateTopic(ctx context.Context, topic *models.Topic) error {
	topic.ID = primitive.NewObjectID()
	topic.CreatedAt = time.Now()
	topic.UpdatedAt = topic.CreatedAt

	_, err := r.topics.InsertOne(ctx, topic)
	return err
}

func (r *MongoRepository) GetTopic(ctx context.Context, id primitive.ObjectID) (*models.Topic, error) {
	var topic models.Topic
	err := r.topics.FindOne(ctx, bson.M{"_id": id}).Decode(&topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *MongoRepository) ListTopics(ctx context.Context, limit, offset int) ([]models.Topic, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "updatedAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.topics.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topics []models.Topic
	if err := cursor.All(ctx, &topics); err != nil {
		return nil, err
	}

	if topics == nil {
		topics = []models.Topic{}
	}

	return topics, nil
}

func (r *MongoRepository) UpdateTopicTimestamp(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.topics.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"updatedAt": time.Now()}},
	)
	return err
}

func (r *MongoRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	message.ID = primitive.NewObjectID()
	message.CreatedAt = time.Now()

	_, err := r.messages.InsertOne(ctx, message)
	return err
}

func (r *MongoRepository) GetMessagesByTopic(ctx context.Context, topicID primitive.ObjectID, limit, offset int) ([]models.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: 1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.messages.Find(ctx, bson.M{"topicId": topicID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	if messages == nil {
		messages = []models.Message{}
	}

	return messages, nil
}

