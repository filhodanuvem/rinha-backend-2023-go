package pessoa

import (
	"log/slog"
	"time"

	"github.com/filhodanuvem/rinha"
	"github.com/filhodanuvem/rinha/internal/config"
	"github.com/google/uuid"
)

func RunWorker(chPessoas chan rinha.Pessoa, chExit chan struct{}, repo *Repository, batch int) {
	slog.Debug("Starting worker...")
	defer slog.Debug("Finishing worker...")
	i := 0
	pessoas := make([]rinha.Pessoa, 0, batch)
	tick := time.NewTicker(config.WorkerTimeout)

	for {
		select {
		case p, ok := <-chPessoas:
			if p.ID != uuid.Nil {
				pessoas = append(pessoas, p)
			}
			if i == batch || !ok {
				if err := repo.Insert(pessoas); err != nil {
					slog.Error(err.Error())
				}
				i = 0
				pessoas = make([]rinha.Pessoa, 0, batch)
			}
			i++

			if !ok {
				chExit <- struct{}{}
				return
			}

		case <-tick.C:
			if len(pessoas) > 0 {
				if err := repo.Insert(pessoas); err != nil {
					slog.Error(err.Error())
				}
			}
			i = 0
			pessoas = make([]rinha.Pessoa, 0, batch)
		}
	}
}
