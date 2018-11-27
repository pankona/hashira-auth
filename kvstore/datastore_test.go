package kvstore

import (
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/datastore"
)

type testEntity struct {
	Value []byte
}

func TestStoreAndLoad(t *testing.T) {
	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, "hashira-auth")
	if err != nil {
		t.Fatalf("failed to create datastore client: %v", err)
	}

	key := datastore.NameKey("testkind", "testkey", nil)
	testStr := "hogehoge"

	buf, err := json.Marshal(testStr)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	e := &testEntity{Value: buf}
	if _, err := dsClient.Put(ctx, key, e); err != nil {
		// TODO: error handling
		t.Fatalf("failed to put: %v", err)
	}

	e2 := testEntity{}
	if err := dsClient.Get(ctx, key, &e2); err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	var ret interface{}
	err = json.Unmarshal(e2.Value, &ret)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if ret != testStr {
		t.Fatalf("unexpected value returned from Get: %v, expected: %v", ret, testStr)
	}
}
