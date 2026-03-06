package qdrant

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strconv"

	"github.com/bryack/obsidian_rag/internal/domain"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

const (
	collectionName = "obsidian_notes"
	VectorSize     = 1024

	filepath   = "file_path"
	hash       = "hash"
	content    = "content"
	headerPath = "header_path"
	links      = "links"
	tags       = "tags"
	project    = "project"
)

var obsidianNamespace = uuid.MustParse("f3f2e850-b5d4-11ef-ac7e-96584d5248b2")

type QdrantStore struct {
	client *qdrant.Client
}

func NewQdrantStore(grpcEndpoint string) (*QdrantStore, error) {
	host, portStr, err := net.SplitHostPort(grpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid grpc endpoint: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q in endpoint: %w", portStr, err)
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &QdrantStore{client: client}, nil
}

func (q *QdrantStore) GetAllHashes(ctx context.Context) (map[string]string, error) {
	if err := q.ensureCollection(ctx); err != nil {
		return nil, err
	}

	hashes := make(map[string]string)

	var offset *qdrant.PointId

	for {
		result, err := q.client.Scroll(ctx, &qdrant.ScrollPoints{
			CollectionName: collectionName,
			WithPayload:    qdrant.NewWithPayloadInclude(filepath, hash),
			Limit:          qdrant.PtrOf(uint32(1000)),
			Offset:         offset,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to scroll points: %w", err)
		}

		for _, p := range result {
			_, okPath := p.Payload[filepath]
			_, okHash := p.Payload[hash]
			if okPath && okHash {
				hashes[getString(p.Payload, filepath)] = getString(p.Payload, hash)
			}
		}
		if len(result) < 1000 {
			break
		}
		offset = result[len(result)-1].Id
	}

	return hashes, nil
}

func (q *QdrantStore) ensureCollection(ctx context.Context) error {
	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		err := q.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
			SparseVectorsConfig: qdrant.NewSparseVectorsConfig(map[string]*qdrant.SparseVectorParams{
				"text": {
					Index: &qdrant.SparseIndexConfig{
						FullScanThreshold: qdrant.PtrOf(uint64(1000)),
					},
				},
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}
	return nil
}

func (q *QdrantStore) Save(ctx context.Context, doc domain.Document) error {
	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         []*qdrant.PointStruct{q.toPoint(doc)},
	})
	return err
}

func (q *QdrantStore) SaveBatch(ctx context.Context, docs []domain.Document) error {
	var points []*qdrant.PointStruct

	for _, doc := range docs {
		points = append(points, q.toPoint(doc))
	}

	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         points,
	})
	return err
}

func (q *QdrantStore) Search(ctx context.Context, vector []float32, sparse map[uint32]float32) ([]domain.Document, error) {
	result, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Prefetch: []*qdrant.PrefetchQuery{
			{
				Query: qdrant.NewQuery(vector...),
				Limit: qdrant.PtrOf(uint64(50)),
			},
			{
				Query: qdrant.NewQuerySparse(prepareSparse(sparse)),
				Using: qdrant.PtrOf("text"),
				Limit: qdrant.PtrOf(uint64(50)),
			},
		},
		Query:       qdrant.NewQueryFusion(qdrant.Fusion_RRF),
		Limit:       qdrant.PtrOf(uint64(10)),
		WithPayload: qdrant.NewWithPayload(true),
		Filter: &qdrant.Filter{
			MustNot: []*qdrant.Condition{
				qdrant.NewMatch(content, ""),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed get result from query: %w", err)
	}

	var docs []domain.Document
	for _, p := range result {
		doc := q.payloadToDocument(p.Payload)
		doc.Score = p.Score
		doc.Vector.SparseVector = extractSparse(p.Vectors)
		docs = append(docs, doc)
	}

	return docs, nil
}

func (q *QdrantStore) payloadToDocument(payload map[string]*qdrant.Value) domain.Document {
	return domain.Document{
		FilePath:   getString(payload, filepath),
		Hash:       getString(payload, hash),
		Content:    getString(payload, content),
		HeaderPath: valueToStrings(payload[headerPath]),
		Metadata: domain.Metadata{
			Tags:    valueToStrings(payload[tags]),
			Project: valueToStrings(payload[project]),
			Links:   valueToStrings(payload[links]),
		},
	}
}

func getString(payload map[string]*qdrant.Value, key string) string {
	if v, ok := payload[key]; ok && v != nil {
		return v.GetStringValue()
	}
	return ""
}

func (q *QdrantStore) toPoint(doc domain.Document) *qdrant.PointStruct {
	data := doc.FilePath + doc.Content
	pointID := uuid.NewSHA1(obsidianNamespace, []byte(data))

	denseVector := qdrant.NewVector(doc.Vector.Dense...)

	sparseVector := qdrant.NewVectorSparse(prepareSparse(doc.Vector.SparseVector))

	return &qdrant.PointStruct{
		Id: qdrant.NewID(pointID.String()),
		Vectors: qdrant.NewVectorsMap(map[string]*qdrant.Vector{
			"":     denseVector,
			"text": sparseVector,
		}),
		Payload: q.documentToPayload(doc),
	}
}

func (q *QdrantStore) documentToPayload(doc domain.Document) map[string]*qdrant.Value {
	return map[string]*qdrant.Value{
		filepath:   qdrant.NewValueString(doc.FilePath),
		hash:       qdrant.NewValueString(doc.Hash),
		content:    qdrant.NewValueString(doc.Content),
		headerPath: NewValueStringList(doc.HeaderPath),
		links:      NewValueStringList(doc.Metadata.Links),
		tags:       NewValueStringList(doc.Metadata.Tags),
		project:    NewValueStringList(doc.Metadata.Project),
	}
}

func prepareSparse(sparse map[uint32]float32) ([]uint32, []float32) {
	var indices []uint32
	var values []float32

	for idx := range sparse {
		indices = append(indices, idx)
	}

	slices.Sort(indices)

	for _, idx := range indices {
		values = append(values, sparse[idx])
	}
	return indices, values
}

func (q *QdrantStore) CountPoints(ctx context.Context) (uint64, error) {
	info, err := q.client.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		return 0, err
	}
	return *info.PointsCount, err
}

func (q *QdrantStore) clear(ctx context.Context) error {
	waitDelete := true
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Wait:           &waitDelete,
		Points:         qdrant.NewPointsSelectorFilter(&qdrant.Filter{}),
	})
	return err
}

func (q *QdrantStore) Get(ctx context.Context, id string) (domain.Document, error) {
	points, err := q.client.Get(ctx, &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids:            []*qdrant.PointId{qdrant.NewID(id)},
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(true),
	})

	if err != nil || len(points) == 0 {
		return domain.Document{}, fmt.Errorf("failed to find points: %w", err)
	}

	p := points[0]
	doc := q.payloadToDocument(p.Payload)

	doc.Vector.SparseVector = extractSparse(p.Vectors)
	return doc, nil
}

func extractSparse(vectors *qdrant.VectorsOutput) map[uint32]float32 {
	if vectors == nil {
		return nil
	}

	named := vectors.GetVectors()
	if named == nil {
		return nil
	}

	vectorMap := named.GetVectors()
	textVector, ok := vectorMap["text"]
	if !ok || textVector == nil {
		return nil
	}

	sparse := textVector.GetSparse()
	if sparse == nil {
		return nil
	}

	res := make(map[uint32]float32)
	for i, idx := range sparse.Indices {
		res[idx] = sparse.Values[i]
	}
	return res
}

func NewValueStringList(strings []string) *qdrant.Value {
	values := make([]*qdrant.Value, len(strings))
	for i, s := range strings {
		values[i] = qdrant.NewValueString(s)
	}

	return &qdrant.Value{
		Kind: &qdrant.Value_ListValue{
			ListValue: &qdrant.ListValue{
				Values: values,
			},
		},
	}
}

func valueToStrings(value *qdrant.Value) []string {
	if value == nil {
		return []string{}
	}

	listValue := value.GetListValue()
	if listValue == nil {
		return []string{}
	}

	if len(listValue.Values) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(listValue.Values))
	for _, v := range listValue.Values {
		str := v.GetStringValue()
		result = append(result, str)
	}

	return result
}
