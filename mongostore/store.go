package mongostore

import (
	"context"
	"fmt"
	"time"

	"github.com/trapck/gopref/cfg"
	"github.com/trapck/gopref/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store is mongo store implementation
type Store struct {
	client *mongo.Client
	db     *mongo.Database
	Pkg    *mongo.Collection
	Usage  *mongo.Collection
}

// Init initializes connetion
func (s *Store) Init() error {
	ctx, cancel := ctx()
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoHost))
	if err != nil {
		return fmt.Errorf("Unable to connect to mongodb: %+v", err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("Mongo ping failed: %+v", err)
	}
	s.client = client

	//TODO: remove line
	client.Database(cfg.MongoDBName).Drop(ctx)

	s.db = client.Database(cfg.MongoDBName)
	err = s.initDocuments(ctx)
	if err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("Error when initializing on of db documents %+v", err)
	}
	return nil
}

// Close closes connetion
func (s *Store) Close() error {
	ctx, cancel := ctx()
	defer cancel()
	return s.client.Disconnect(ctx)
}

// SavePkgs packages
func (s *Store) SavePkgs(p []model.Pkg) error {
	ctx, cancel := ctx()
	defer cancel()
	docs := []interface{}{}
	for _, v := range p {
		docs = append(docs, v)
	}
	_, err := s.Pkg.InsertMany(ctx, docs)
	return err
}

// SaveUsages usages
func (s *Store) SaveUsages(u []model.PkgImportUsage) error {
	ctx, cancel := ctx()
	defer cancel()
	docs := []interface{}{}
	for _, v := range u {
		docs = append(docs, v)
	}
	_, err := s.Usage.InsertMany(ctx, docs)
	return err
}

// GetCombinedUsages returns combined import usages
func (s *Store) GetCombinedUsages() ([]model.CombinedUsage, error) {
	ctx, cancel := ctx()
	defer cancel()
	cur, err := s.Usage.Aggregate(ctx, mongo.Pipeline{bson.D{
		{
			"$group",
			bson.D{
				{"_id", bson.A{"$path", "$inFilePath"}},
				{"Path", bson.D{{"$first", "$path"}}},
				{"InFile", bson.D{{"$first", "$inFilePath"}}},
				{"Text", bson.D{{"$addToSet", "$text"}}},
			},
		},
	}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []model.CombinedUsage
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Store) initDocuments(ctx context.Context) error {
	s.Pkg = s.db.Collection("pkg")
	s.Usage = s.db.Collection("usages")
	return nil
}

func ctx(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := cfg.MongoDefTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(context.Background(), t)
}
