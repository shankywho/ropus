package memory

import (
	"math"
	"sort"
	"sync"
	"time"
)

// CaseEmbeddingRecord represents a vector-indexed fraud case.
type CaseEmbeddingRecord struct {
	CaseID        string    `json:"case_id"`
	Title         string    `json:"title"`
	AttackPattern string    `json:"attack_pattern"`
	Resolution    string    `json:"resolution"`
	Embedding     []float64 `json:"-"`
	IndexedAt     time.Time `json:"indexed_at"`
}

// ScoredCaseResult represents a semantic search match with cosine similarity score.
type ScoredCaseResult struct {
	Case      *CaseEmbeddingRecord `json:"case"`
	Similarity float64             `json:"similarity"` // 0.0 to 1.0
}

// VectorStore provides vector storage and nearest-neighbor search for RAG.
type VectorStore struct {
	mu      sync.RWMutex
	records []*CaseEmbeddingRecord
}

// NewVectorStore initializes the in-memory/pgvector store.
func NewVectorStore() *VectorStore {
	vs := &VectorStore{
		records: make([]*CaseEmbeddingRecord, 0),
	}
	// Seed historical case patterns for RAG
	vs.IndexCase("case_hist_01", "Account Takeover via SIM Swap", "ATO_SIM_SWAP", "Challenged WebAuthn, froze account", []float64{0.92, 0.12, 0.45, 0.88})
	vs.IndexCase("case_hist_02", "Synthetic Identity Farm Creation", "SYNTHETIC_IDENTITY", "Blocked cluster of 14 accounts", []float64{0.15, 0.95, 0.82, 0.25})
	vs.IndexCase("case_hist_03", "Distributed Carding Wave", "CARDING_BURST", "Isolated proxy subnet, notified consortium", []float64{0.85, 0.44, 0.91, 0.77})
	return vs
}

// IndexCase adds a case embedding into the vector database.
func (v *VectorStore) IndexCase(caseID, title, pattern, resolution string, emb []float64) *CaseEmbeddingRecord {
	v.mu.Lock()
	defer v.mu.Unlock()

	rec := &CaseEmbeddingRecord{
		CaseID:        caseID,
		Title:         title,
		AttackPattern: pattern,
		Resolution:    resolution,
		Embedding:     emb,
		IndexedAt:     time.Now().UTC(),
	}
	v.records = append(v.records, rec)
	return rec
}

// SearchSimilarCases performs cosine similarity ranking.
func (v *VectorStore) SearchSimilarCases(queryEmb []float64, topK int) []*ScoredCaseResult {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var scored []*ScoredCaseResult
	for _, rec := range v.records {
		sim := cosineSimilarity(queryEmb, rec.Embedding)
		scored = append(scored, &ScoredCaseResult{
			Case:       rec,
			Similarity: math.Round(sim*1000) / 1000.0,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Similarity > scored[j].Similarity
	})

	if topK > len(scored) {
		topK = len(scored)
	}
	return scored[:topK]
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}
	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
