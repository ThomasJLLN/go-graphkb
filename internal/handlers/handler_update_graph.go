package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/client"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"golang.org/x/sync/semaphore"
)

func handleUpdate(registry sources.Registry, fn func(source string, body io.Reader) error, sem *semaphore.Weighted) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := IsTokenValid(registry, r)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		if !ok {
			ReplyWithUnauthorized(w)
			return
		}

		{
			ok = sem.TryAcquire(1)
			if !ok {
				ReplyWithTooManyRequests(w)
				return
			}
			defer sem.Release(1)

			if err = fn(source, r.Body); err != nil {
				ReplyWithInternalError(w, err)
				return
			}
		}

		_, err = bytes.NewBufferString("Graph has been received and will be processed soon").WriteTo(w)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}
	}
}

// PutSchema upsert an asset into the graph of the data source
func PutSchema(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphSchemaRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpdateSchema(source, requestBody.Schema)
		return nil
	}, sem)
}

// PutAssets upsert assets into the graph of the data source
func PutAssets(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpsertAssets(source, requestBody.Assets)
		return nil
	}, sem)
}

// PutRelations upsert relations into the graph of the data source
func PutRelations(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpsertRelations(source, requestBody.Relations)
		return nil
	}, sem)
}

// DeleteAssets delete assets from the graph of the data source
func DeleteAssets(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.RemoveAssets(source, requestBody.Assets)
		return nil
	}, sem)
}

// DeleteRelations upsert relation into the graph of the data source
func DeleteRelations(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.RemoveRelations(source, requestBody.Relations)
		return nil
	}, sem)
}
