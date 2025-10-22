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
	"github.com/redis/go-redis/v9"
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
	}
	elapsed := time.Since(start)
	if elapsed < 10*time.Millisecond {
		time.Sleep(10*time.Millisecond - elapsed)
	}
}

func (r *RawDataHandler) rawdataHandler(data model.RawDataWiki) {
	result := data.LinksRes

	// Ensure the source title exists
	if data.TitleQ.ID == "" {
		data.TitleQ.ID = uuid.New().String()

		// Check Redis cache before inserting
		cacheKey := "title:" + data.TitleQ.Title
		exists, _ := r.store.RedisClient.GetClient().Exists(context.Background(), cacheKey).Result()

		if exists == 0 {
			if err := r.store.TitlesTable.Insert(map[string]interface{}{
				"title_id": data.TitleQ.ID,
				"name":     data.TitleQ.Title,
			}); err == nil {
				// Cache the title
				r.store.RedisClient.GetClient().Set(context.Background(), cacheKey, data.TitleQ.ID, 0)
			} else {
				log.Printf("[RawDataHandler] Failed to insert title %s: %v", data.TitleQ.Title, err)
				return
			}
		} else {
			id, _ := r.store.RedisClient.GetClient().Get(context.Background(), cacheKey).Result()
			data.TitleQ.ID = id
		}
	}

	// Process linked titles
	for _, page := range result.Query.Pages {
		for _, link := range page.Links {
			if link.Ns != 0 {
				continue
			}

			titleCacheKey := "title:" + link.Title

			// Check Redis for existing title ID
			titleID, err := r.store.RedisClient.GetClient().Get(context.Background(), titleCacheKey).Result()
			if err == redis.Nil {
				titleID = uuid.New().String()
				if err := r.store.TitlesTable.Insert(map[string]interface{}{
					"title_id": titleID,
					"name":     link.Title,
				}); err != nil {
					log.Printf("[RawDataHandler] Failed to insert title '%s': %v", link.Title, err)
					continue
				}
				// Save to cache
				r.store.RedisClient.GetClient().Set(context.Background(), titleCacheKey, titleID, 0)
			} else if err != nil {
				log.Printf("[RawDataHandler] Redis error: %v", err)
				continue
			}

			// Build and cache relationship
			pairCacheKey := fmt.Sprintf("pair:%s:%s", data.TitleQ.ID, titleID)
			exists, _ := r.store.RedisClient.GetClient().Exists(context.Background(), pairCacheKey).Result()
			if exists == 0 {
				pairID := uuid.New().String()
				if err := r.store.PairsTable.Insert(map[string]interface{}{
					"pair_id":   pairID,
					"title_src": data.TitleQ.ID,
					"title_dst": titleID,
				}); err != nil {
					log.Printf("[RawDataHandler] Failed to insert pair (%s → %s): %v", data.TitleQ.Title, link.Title, err)
					continue
				}
				// Cache the pair
				r.store.RedisClient.GetClient().Set(context.Background(), pairCacheKey, pairID, 0)
			}
		}
	}
}
