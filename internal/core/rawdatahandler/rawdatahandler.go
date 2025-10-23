package rawdatahandler

import (
	"context"
	"fmt"
	"log"
	"time"
	"wikicrawler/internal/infra"
	"wikicrawler/internal/model"
	"wikicrawler/internal/utils/processor"
	"wikicrawler/internal/utils/workers"
	tasks "wikicrawler/internal/utils/workers/task"

	"github.com/google/uuid"
)

type RawDataHandler struct {
	processor.BaseProcessor
	workerPool *workers.WorkerPool
	store      *infra.WikiStore
}

func NewRawDataHandler(store *infra.WikiStore, nworkers, ntasks int) *RawDataHandler {
	r := &RawDataHandler{}
	r.workerPool = workers.NewWorkerPool(nworkers, ntasks)
	r.store = store
	r.Init(r)
	return r
}

func (r *RawDataHandler) Start() error {
	r.workerPool.Start()
	if err := r.BaseProcessor.Start(); err != nil {
		return err
	}
	return nil
}

func (r *RawDataHandler) Stop() error {
	r.workerPool.Stop()
	if err := r.BaseProcessor.Stop(); err != nil {
		return err
	}
	return nil
}

func (r *RawDataHandler) RunningTask() {
	start := time.Now()
	data := <-r.store.RawDataQ
	task := tasks.NewTask(r.rawdataHandler, data)
	ok := r.workerPool.Tasks.Push(task)
	if !ok {
		fmt.Printf("[RawDataHandler] failed to queue data handle task\n")
		time.Sleep(10 * time.Millisecond)
	}
	elapsed := time.Since(start)
	if elapsed < 2*time.Millisecond {
		time.Sleep(2*time.Millisecond - elapsed)
	}
}
func (r *RawDataHandler) rawdataHandler(data model.RawDataWiki) {
	ctx := context.Background()
	result := data.LinksRes
	client := r.store.RedisClient.GetClient()
	// --- Ensure source title exists (atomic with SETNX) ---
	if data.TitleQ.ID == "" {
		data.TitleQ.ID = uuid.New().String()
		cacheKey := "title:" + data.TitleQ.Title

		setOK, err := client.
			SetNX(ctx, cacheKey, data.TitleQ.ID, 0).Result()
		if err != nil {
			log.Printf("[RawDataHandler] Redis SETNX error for %s: %v", cacheKey, err)
			return
		}

		if setOK {
			// Key was new → safe to insert into DB
			if err := r.store.TitlesTable.Insert(map[string]interface{}{
				"title_id": data.TitleQ.ID,
				"name":     data.TitleQ.Title,
			}); err != nil {
				log.Printf("[RawDataHandler] Failed to insert title %s: %v", data.TitleQ.Title, err)
				return
			}
		} else {
			// Key existed → get the stored ID
			id, _ := client.Get(ctx, cacheKey).Result()
			data.TitleQ.ID = id
		}
	}

	// --- Process linked titles ---
	for _, page := range result.Query.Pages {
		for _, link := range page.Links {
			if link.Ns != 0 {
				continue
			}

			titleCacheKey := "title:" + link.Title
			titleID := uuid.New().String()

			// Atomic create title if not exists
			setOK, err := client.
				SetNX(ctx, titleCacheKey, titleID, 0).Result()
			if err != nil {
				log.Printf("[RawDataHandler] Redis SETNX error: %v", err)
				continue
			}

			if setOK {
				if err := r.store.TitlesTable.Insert(map[string]interface{}{
					"title_id": titleID,
					"name":     link.Title,
				}); err != nil {
					log.Printf("[RawDataHandler] Failed to insert title '%s': %v", link.Title, err)
					continue
				}

				// push to title query queue for new request
				r.store.TitlsToQueryQ <- model.TitleQuery{
					Title: link.Title,
					ID:    titleID,
				}

			} else {
				// Already exists → use cached ID
				id, _ := client.Get(ctx, titleCacheKey).Result()
				titleID = id
			}

			// --- Build relationship (also atomic) ---
			pairCacheKey := fmt.Sprintf("pair:%s:%s", data.TitleQ.ID, titleID)
			pairID := uuid.New().String()

			setOK, err = client.
				SetNX(ctx, pairCacheKey, pairID, 0).Result()
			if err != nil {
				log.Printf("[RawDataHandler] Redis SETNX error (pair): %v", err)
				continue
			}

			if setOK {
				if err := r.store.PairsTable.Insert(map[string]interface{}{
					"pair_id":   pairID,
					"title_src": data.TitleQ.ID,
					"title_dst": titleID,
				}); err != nil {
					log.Printf("[RawDataHandler] Failed to insert pair (%s → %s): %v",
						data.TitleQ.Title, link.Title, err)
					continue
				}
			}
		}
	}
}
