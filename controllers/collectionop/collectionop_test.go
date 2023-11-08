package collectionop

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCollectionController_GetListofCollectionData(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.GetListofCollectionData(tt.args.ctx)
		})
	}
}

func TestCollectionController_GetDetailCollectionData(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.GetDetailCollectionData(tt.args.ctx)
		})
	}
}

func TestCollectionController_GetDetailCollectionDatabyID(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.GetDetailCollectionDatabyID(tt.args.ctx)
		})
	}
}

func TestCollectionController_GetDetailCollectionDatabyName(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.GetDetailCollectionDatabyName(tt.args.ctx)
		})
	}
}

func TestCollectionController_UpdateCollectionData(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.UpdateCollectionData(tt.args.ctx)
		})
	}
}

func TestCollectionController_DeleteCollectionDatabyID(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DeleteCollectionDatabyID(tt.args.ctx)
		})
	}
}

func TestCollectionController_buildProjectionFromJSON(t *testing.T) {
	type args struct {
		jsonData    []byte
		converttype string
	}
	tests := []struct {
		name    string
		c       *CollectionController
		args    args
		want    bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.buildProjectionFromJSON(tt.args.jsonData, tt.args.converttype)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectionController.buildProjectionFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectionController.buildProjectionFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionController_buildProjection(t *testing.T) {
	type args struct {
		jsonMap     map[string]interface{}
		prefix      string
		projection  bson.M
		converttype string
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.buildProjection(tt.args.jsonMap, tt.args.prefix, tt.args.projection, tt.args.converttype)
		})
	}
}

func TestCollectionController_CollectionObjectRevision(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		c    *CollectionController
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.CollectionObjectRevision(tt.args.ctx)
		})
	}
}
