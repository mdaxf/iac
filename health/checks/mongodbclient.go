package checks

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClientCheck struct {
	Ctx           context.Context
	MongoDBClient *mongo.Client
}

func NewMongoClientCheck(ctx context.Context, client *mongo.Client) MongoClientCheck {
	return MongoClientCheck{
		Ctx:           ctx,
		MongoDBClient: client,
	}
}

func CheckMongoClientStatus(ctx context.Context, client *mongo.Client) error {
	check := NewMongoClientCheck(ctx, client)
	return check.CheckStatus()
}

func (check MongoClientCheck) CheckStatus() error {
	ctx := check.Ctx
	err := check.MongoDBClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("mongoDB health check failed on ping: %w", err)
	}

	return nil
}
