package firestoretest

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/tj/assert"
)

func CreateTestClient(ctx context.Context, t *testing.T) *firestore.Client {
	t.Helper()

	client, err := firestore.NewClient(ctx, "servizio-be")
	assert.NoError(t, err)

	return client
}

func DeleteCollection(ctx context.Context, t *testing.T, fsClient *firestore.Client, collection string) {
	documentRefs, err := fsClient.Collection(collection).DocumentRefs(ctx).GetAll()
	assert.NoError(t, err)

	for _, document := range documentRefs {
		_, err := document.Delete(ctx)
		assert.NoError(t, err)
	}
}
