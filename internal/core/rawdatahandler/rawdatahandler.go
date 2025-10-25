package rawdatahandler

import (
	"fmt"
	"log"
	"strings"
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
		fmt.Printf("[RawDataHandler] failed to queue data handle task %d\n", r.workerPool.Tasks.Size())
		time.Sleep(10 * time.Millisecond)
	}
	elapsed := time.Since(start)
	if elapsed < 2*time.Millisecond {
		time.Sleep(2*time.Millisecond - elapsed)
	}
}
func (r *RawDataHandler) rawdataHandler(data model.RawDataWiki) {
	result := data.LinksRes
	// --- Ensure source title exists (atomic with SETNX) ---
	if data.TitleQ.ID == "" {
		data.TitleQ.ID = uuid.New().String()
		if err := r.store.TitlesTable.Insert(map[string]interface{}{
			"title_id": data.TitleQ.ID,
			"name":     data.TitleQ.Title,
		}); err != nil {
			log.Printf("[RawDataHandler] Failed to insert title %s: %v", data.TitleQ.Title, err)
			return
		} else {
			log.Printf("[RawDataHandler] Success to insert title '%s' to titles", data.TitleQ.Title)

		}
	}

	// --- Process linked titles ---
	for _, page := range result.Query.Pages {
		for _, link := range page.Links {
			if link.Ns != 0 {
				continue
			}

			titleID := uuid.New().String()

			if err := r.store.TitlesTable.Insert(map[string]interface{}{
				"title_id": titleID,
				"name":     link.Title,
			}); err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					// Title already exists → fetch its ID
					existing, getErr := r.store.TitlesTable.GetRecordByKey("name", link.Title)
					if getErr != nil {
						log.Printf("[RawDataHandler] Failed to get existing title '%s': %v", link.Title, getErr)
						continue
					}

					if id, ok := existing["title_id"].(string); ok {
						titleID = id
						log.Printf("[RawDataHandler] Found existing title '%s' with ID %s", link.Title, titleID)
					} else {
						log.Printf("[RawDataHandler] Invalid data for existing title '%s'", link.Title)
						continue
					}
				} else {
					log.Printf("[RawDataHandler] Failed to insert title '%s': %v", link.Title, err)
					continue
				}
			} else {
				log.Printf("[RawDataHandler] Success to insert title '%s' to titles", link.Title)
				r.store.TitlsToQueryQ <- model.TitleQuery{Title: link.Title, ID: titleID}
			}

			// --- Build relationship (also atomic) ---
			pairID := uuid.New().String()

			if err := r.store.PairsTable.Insert(map[string]interface{}{
				"pair_id":   pairID,
				"title_src": data.TitleQ.ID,
				"title_dst": titleID,
			}); err != nil {
				log.Printf("[RawDataHandler] Failed to insert pair (%s → %s): %v",
					data.TitleQ.Title, link.Title, err)
				continue
			} else {
				log.Printf("[RawDataHandler] Success to insert pair (%s → %s) to pairs",
					data.TitleQ.Title, link.Title)
			}
		}
	}
}
