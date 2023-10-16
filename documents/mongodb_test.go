package documents

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestDocDB_ConnectMongoDB(t *testing.T) {
	tests := []struct {
		name    string
		doc     *DocDB
		want    *DocDB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.ConnectMongoDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.ConnectMongoDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.ConnectMongoDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocDB_QueryCollection(t *testing.T) {
	type args struct {
		collectionname string
		filter         bson.M
		projection     bson.M
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		want    []bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.QueryCollection(tt.args.collectionname, tt.args.filter, tt.args.projection)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.QueryCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.QueryCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocDB_GetDefaultItembyName(t *testing.T) {
	type args struct {
		collectionname string
		name           string
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		want    bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.GetDefaultItembyName(tt.args.collectionname, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.GetDefaultItembyName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.GetDefaultItembyName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocDB_GetItembyID(t *testing.T) {
	type args struct {
		collectionname string
		id             string
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		want    bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.GetItembyID(tt.args.collectionname, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.GetItembyID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.GetItembyID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocDB_UpdateCollection(t *testing.T) {
	type args struct {
		collectionname string
		filter         bson.M
		update         bson.M
		idata          interface{}
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.doc.UpdateCollection(tt.args.collectionname, tt.args.filter, tt.args.update, tt.args.idata); (err != nil) != tt.wantErr {
				t.Errorf("DocDB.UpdateCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocDB_InsertCollection(t *testing.T) {
	type args struct {
		collectionname string
		idata          interface{}
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		want    *mongo.InsertOneResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.InsertCollection(tt.args.collectionname, tt.args.idata)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.InsertCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.InsertCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocDB_DeleteItemFromCollection(t *testing.T) {
	type args struct {
		collectionname string
		documentid     string
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.doc.DeleteItemFromCollection(tt.args.collectionname, tt.args.documentid); (err != nil) != tt.wantErr {
				t.Errorf("DocDB.DeleteItemFromCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocDB_convertToBsonM(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		doc     *DocDB
		args    args
		want    bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.convertToBsonM(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DocDB.convertToBsonM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DocDB.convertToBsonM() = %v, want %v", got, tt.want)
			}
		})
	}
}
